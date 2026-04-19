package usermap

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/go-redis/redis/v8"
	database "github.com/raiashpanda007/rivon/engine/internals/Database"
)

const AdminID = "00000000-0000-0000-0000-000000000001"

type TxnType string

const (
	Debit  TxnType = "DEBIT"
	Credit TxnType = "CREDIT"
)

type redisMessageAssetStruct struct {
	MarketID string `json:"marketId"`
	Quantity int    `json:"quantity"`
}

type wallet struct {
	Balance int `json:"balance"`
}

type redisMessageStruct struct {
	UserId string                    `json:"id"`
	Wallet wallet                    `json:"userWallet"`
	Assets []redisMessageAssetStruct `json:"assets"`
}

type asset struct {
	Quantity int `json:"quantity"`
}

type UserAssetsStruct struct {
	Assets map[string]asset
	mutex  sync.Mutex
}

type UserWalletStruct struct {
	Wallet wallet
	mutex  sync.Mutex
}

type UserWallet struct {
	mu                   sync.RWMutex
	WalletMap            map[string]*UserWalletStruct
	AssetMap             map[string]*UserAssetsStruct
	AdminWallet          *UserWalletStruct
	AdminAssets          *UserAssetsStruct
	adminEscrowBalance   int            // funds held from user BUY orders
	adminOwnLocked       int            // admin's own pending BUY commitments
	adminEscrowAssets    map[string]int // assets held from user SELL orders per market
	adminOwnLockedAssets map[string]int // admin's own pending SELL commitments per market
	redisClient          *redis.Client
	ctx                  context.Context
}

//////////////////// INIT ////////////////////

func InitUserMap(redisClient *redis.Client, ctx context.Context, db *database.Database) (*UserWallet, error) {
	uw := &UserWallet{
		WalletMap:            make(map[string]*UserWalletStruct),
		AssetMap:             make(map[string]*UserAssetsStruct),
		adminEscrowAssets:    make(map[string]int),
		adminOwnLockedAssets: make(map[string]int),
		redisClient:          redisClient,
		ctx:                  ctx,
	}

	if err := uw.loadAdminFromDB(db); err != nil {
		return nil, err
	}

	uw.AdminWallet = uw.WalletMap[AdminID]
	uw.AdminAssets = uw.AssetMap[AdminID]

	return uw, nil
}

func (r *UserWallet) loadAdminFromDB(db *database.Database) error {
	data, err := db.GetAdminData(AdminID)
	if err != nil {
		return errors.New("failed to load admin data from database: " + err.Error())
	}

	assetMap := make(map[string]asset)
	for _, a := range data.Assets {
		assetMap[a.MarketID] = asset{Quantity: a.Quantity}
		r.adminEscrowAssets[a.MarketID] = a.LockedQty
	}

	r.WalletMap[AdminID] = &UserWalletStruct{
		Wallet: wallet{Balance: data.Balance},
	}
	r.AssetMap[AdminID] = &UserAssetsStruct{
		Assets: assetMap,
	}
	r.adminEscrowBalance = data.LockedBalance
	return nil
}

//////////////////// LOAD USER ////////////////////

func (r *UserWallet) addUserWallet(userID string) error {
	val, err := r.redisClient.Get(r.ctx, userID).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New("can't load your wallet, please try again later")
		}
		return err
	}

	var user redisMessageStruct
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return err
	}

	r.WalletMap[userID] = &UserWalletStruct{
		Wallet: user.Wallet,
	}

	assetMap := make(map[string]asset)
	for _, a := range user.Assets {
		assetMap[a.MarketID] = asset{Quantity: a.Quantity}
	}

	r.AssetMap[userID] = &UserAssetsStruct{
		Assets: assetMap,
	}

	return nil
}

//////////////////// GET USER ////////////////////

func (r *UserWallet) GetUser(userId string) (*UserWalletStruct, *UserAssetsStruct, error) {
	r.mu.RLock()
	w, ok1 := r.WalletMap[userId]
	a, ok2 := r.AssetMap[userId]
	r.mu.RUnlock()

	if ok1 && ok2 {
		return w, a, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.WalletMap[userId]; !ok {
		if err := r.addUserWallet(userId); err != nil {
			return nil, nil, err
		}
	}

	return r.WalletMap[userId], r.AssetMap[userId], nil
}

//////////////////// WALLET OPS ////////////////////

func (w *UserWalletStruct) Add(amount int) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.Wallet.Balance += amount
}

func (w *UserWalletStruct) Sub(amount int) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.Wallet.Balance < amount {
		return errors.New("you don't have enough funds for this order")
	}
	w.Wallet.Balance -= amount
	return nil
}

// SubEscrow releases escrowed funds — no balance check (escrow is guaranteed to exist).
func (w *UserWalletStruct) SubEscrow(amount int) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.Wallet.Balance -= amount
}

//////////////////// ASSET OPS ////////////////////

func (a *UserAssetsStruct) Add(marketId string, qty int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	val := a.Assets[marketId]
	val.Quantity += qty
	a.Assets[marketId] = val
}

func (a *UserAssetsStruct) Sub(marketId string, qty int) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	val, ok := a.Assets[marketId]
	if !ok || val.Quantity < qty {
		return errors.New("you don't have enough assets to place this sell order")
	}

	val.Quantity -= qty
	a.Assets[marketId] = val
	return nil
}

// SubEscrowAsset releases escrowed assets — no quantity check (escrow is guaranteed to exist).
func (a *UserAssetsStruct) SubEscrowAsset(marketId string, qty int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	val := a.Assets[marketId]
	val.Quantity -= qty
	a.Assets[marketId] = val
}

//////////////////// FLOW ////////////////////

// Step 1: Lock money (BUY order)
func (r *UserWallet) LockMoney(userId string, amount int) error {
	if userId == AdminID {
		r.mu.Lock()
		defer r.mu.Unlock()
		available := r.AdminWallet.Wallet.Balance - r.adminEscrowBalance - r.adminOwnLocked
		if available < amount {
			return errors.New("you don't have enough funds for this order")
		}
		r.adminOwnLocked += amount
		return nil
	}

	userWallet, _, err := r.GetUser(userId)
	if err != nil {
		return err
	}
	if err := userWallet.Sub(amount); err != nil {
		return err
	}
	r.mu.Lock()
	r.AdminWallet.Add(amount)
	r.adminEscrowBalance += amount
	r.mu.Unlock()
	return nil
}

// Step 1b: Unlock money (BUY order cancelled)
func (r *UserWallet) UnlockMoney(userId string, amount int) error {
	if userId == AdminID {
		r.mu.Lock()
		r.adminOwnLocked -= amount
		r.mu.Unlock()
		return nil
	}

	r.mu.Lock()
	r.AdminWallet.SubEscrow(amount)
	r.adminEscrowBalance -= amount
	r.mu.Unlock()

	userWallet, _, err := r.GetUser(userId)
	if err != nil {
		return err
	}
	userWallet.Add(amount)
	return nil
}

// Step 2: Lock asset (SELL order)
func (r *UserWallet) LockAsset(userId, marketId string, qty int) error {
	if userId == AdminID {
		r.mu.Lock()
		defer r.mu.Unlock()
		available := r.AdminAssets.Assets[marketId].Quantity - r.adminEscrowAssets[marketId] - r.adminOwnLockedAssets[marketId]
		if available < qty {
			return errors.New("you don't have enough assets to place this sell order")
		}
		r.adminOwnLockedAssets[marketId] += qty
		return nil
	}

	_, assets, err := r.GetUser(userId)
	if err != nil {
		return err
	}
	if err := assets.Sub(marketId, qty); err != nil {
		return err
	}
	r.mu.Lock()
	r.AdminAssets.Add(marketId, qty)
	r.adminEscrowAssets[marketId] += qty
	r.mu.Unlock()
	return nil
}

// Step 2b: Unlock asset (SELL order cancelled)
func (r *UserWallet) UnlockAsset(userId, marketId string, qty int) error {
	if userId == AdminID {
		r.mu.Lock()
		r.adminOwnLockedAssets[marketId] -= qty
		r.mu.Unlock()
		return nil
	}

	r.mu.Lock()
	r.AdminAssets.SubEscrowAsset(marketId, qty)
	r.adminEscrowAssets[marketId] -= qty
	r.mu.Unlock()

	_, assets, err := r.GetUser(userId)
	if err != nil {
		return err
	}
	assets.Add(marketId, qty)
	return nil
}

// Step 3: Execute Trade
func (r *UserWallet) ExecuteTrade(buyerId, sellerId, marketId string, qty, price int) error {
	total := qty * price

	_, buyerAssets, _ := r.GetUser(buyerId)
	sellerWallet, _, _ := r.GetUser(sellerId)

	r.mu.Lock()
	defer r.mu.Unlock()

	switch {
	case buyerId == AdminID:
		// Admin's own BUY filled: pay seller from admin's own committed funds.
		if err := r.AdminWallet.Sub(total); err != nil {
			return err
		}
		r.adminOwnLocked -= total
		sellerWallet.Add(total)
		r.AdminAssets.Add(marketId, qty)

	case sellerId == AdminID:
		// Admin's own SELL filled: release asset, collect buyer's escrowed payment.
		r.AdminAssets.SubEscrowAsset(marketId, qty)
		r.adminOwnLockedAssets[marketId] -= qty
		buyerAssets.Add(marketId, qty)
		r.AdminWallet.SubEscrow(total)
		r.adminEscrowBalance -= total
		r.AdminWallet.Add(total) // net: Balance unchanged, escrow drops; admin retains funds

	default:
		// Normal trade: release buyer's escrowed money to seller, seller's escrowed asset to buyer.
		r.AdminWallet.SubEscrow(total)
		r.adminEscrowBalance -= total
		sellerWallet.Add(total)
		r.AdminAssets.SubEscrowAsset(marketId, qty)
		r.adminEscrowAssets[marketId] -= qty
		buyerAssets.Add(marketId, qty)
	}

	return nil
}

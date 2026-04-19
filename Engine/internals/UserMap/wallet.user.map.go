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
	mu          sync.RWMutex
	WalletMap   map[string]*UserWalletStruct
	AssetMap    map[string]*UserAssetsStruct
	AdminWallet *UserWalletStruct
	AdminAssets *UserAssetsStruct
	redisClient *redis.Client
	ctx         context.Context
}

//////////////////// INIT ////////////////////

func InitUserMap(redisClient *redis.Client, ctx context.Context, db *database.Database) (*UserWallet, error) {
	uw := &UserWallet{
		WalletMap:   make(map[string]*UserWalletStruct),
		AssetMap:    make(map[string]*UserAssetsStruct),
		redisClient: redisClient,
		ctx:         ctx,
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
	}

	r.WalletMap[AdminID] = &UserWalletStruct{
		Wallet: wallet{Balance: data.Balance},
	}
	r.AssetMap[AdminID] = &UserAssetsStruct{
		Assets: assetMap,
	}
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

//////////////////// FLOW ////////////////////

// Step 1: Lock money (BUY order)
func (r *UserWallet) LockMoney(userId string, amount int) error {
	userWallet, _, err := r.GetUser(userId)
	if err != nil {
		return err
	}

	if err := userWallet.Sub(amount); err != nil {
		return err
	}

	r.AdminWallet.Add(amount)
	return nil
}

// Step 1b: Unlock money (BUY order cancelled)
func (r *UserWallet) UnlockMoney(userId string, amount int) error {
	if err := r.AdminWallet.Sub(amount); err != nil {
		return err
	}
	userWallet, _, err := r.GetUser(userId)
	if err != nil {
		return err
	}
	userWallet.Add(amount)
	return nil
}

// Step 2b: Unlock asset (SELL order cancelled)
func (r *UserWallet) UnlockAsset(userId, marketId string, qty int) error {
	if err := r.AdminAssets.Sub(marketId, qty); err != nil {
		return err
	}
	_, assets, err := r.GetUser(userId)
	if err != nil {
		return err
	}
	assets.Add(marketId, qty)
	return nil
}

// Step 2: Lock asset (SELL order)
func (r *UserWallet) LockAsset(userId, marketId string, qty int) error {
	_, assets, err := r.GetUser(userId)
	if err != nil {
		return err
	}

	if err := assets.Sub(marketId, qty); err != nil {
		return err
	}

	r.AdminAssets.Add(marketId, qty)
	return nil
}

// Step 3: Execute Trade
func (r *UserWallet) ExecuteTrade(buyerId, sellerId, marketId string, qty, price int) error {
	total := qty * price

	_, buyerAssets, _ := r.GetUser(buyerId)
	sellerWallet, _, _ := r.GetUser(sellerId)

	// Admin -> Seller (money)
	if err := r.AdminWallet.Sub(total); err != nil {
		return err
	}
	sellerWallet.Add(total)

	// Admin -> Buyer (asset)
	if err := r.AdminAssets.Sub(marketId, qty); err != nil {
		return err
	}
	buyerAssets.Add(marketId, qty)

	return nil
}

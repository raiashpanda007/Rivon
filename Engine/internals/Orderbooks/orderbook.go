package orderbooks

import (
	"errors"
	"log"
	"log/slog"

	"github.com/google/uuid"
	heap "github.com/raiashpanda007/rivon/engine/internals/utils"
)

type OrderSide string

const (
	BUY  OrderSide = "BUY"
	SELL OrderSide = "SELL"
)

type Order struct {
	Id       string
	Quantity int
	Filled   int
	Side     OrderSide
	UserId   string
	Price    int
	StreamId string
}

type Fills struct {
	Price       int
	Quantity    int
	TradeId     string
	OtherUserId string
	OrderId     string
}

type OrderBook struct {
	Bids         map[int][]*Order
	Asks         map[int][]*Order
	BidHeap      *heap.MaxHeap
	AskHeap      *heap.MinHeap
	UserOrderMap map[string]map[string]*Order
	CurrentPrice int
	LastTradeId  string
	LastOrderId  string
	LastStreamId string
}

func NewOrderBook(lastTradeId, lastOrderId, lastStreamId string, bids map[int][]Order, asks map[int][]Order, askHeap *heap.MinHeap, bidHeap *heap.MaxHeap, currentPrice int) OrderBook {
	bidsCopy := make(map[int][]*Order)
	asksCopy := make(map[int][]*Order)
	userOrderMap := make(map[string]map[string]*Order)

	for price, orders := range bids {
		for i := range orders {
			orderPtr := &orders[i]

			bidsCopy[price] = append(bidsCopy[price], orderPtr)

			if _, ok := userOrderMap[orderPtr.UserId]; !ok {
				userOrderMap[orderPtr.UserId] = make(map[string]*Order)
			}
			userOrderMap[orderPtr.UserId][orderPtr.Id] = orderPtr

		}
	}

	for price, orders := range asks {
		for i := range orders {
			orderPtr := &orders[i]

			asksCopy[price] = append(asksCopy[price], orderPtr)

			if _, ok := userOrderMap[orderPtr.UserId]; !ok {
				userOrderMap[orderPtr.UserId] = make(map[string]*Order)
			}
			userOrderMap[orderPtr.UserId][orderPtr.Id] = orderPtr
		}
	}

	return OrderBook{
		Bids:         bidsCopy,
		Asks:         asksCopy,
		BidHeap:      bidHeap,
		AskHeap:      askHeap,
		UserOrderMap: userOrderMap,
		CurrentPrice: currentPrice,
		LastTradeId:  lastTradeId,
		LastOrderId:  lastOrderId,
		LastStreamId: lastStreamId,
	}
}

func (r *OrderBook) GetSnapshot() OrderBook {
	// Map from original order pointer → its copy, so shared pointers
	// between Bids/Asks and UserOrderMap all point to the same new object.
	orderCopy := make(map[*Order]*Order)
	copyOrder := func(o *Order) *Order {
		if _, ok := orderCopy[o]; !ok {
			cp := *o
			orderCopy[o] = &cp
		}
		return orderCopy[o]
	}

	bidsCopy := make(map[int][]*Order, len(r.Bids))
	for price, orders := range r.Bids {
		slice := make([]*Order, len(orders))
		for i, o := range orders {
			slice[i] = copyOrder(o)
		}
		bidsCopy[price] = slice
	}

	asksCopy := make(map[int][]*Order, len(r.Asks))
	for price, orders := range r.Asks {
		slice := make([]*Order, len(orders))
		for i, o := range orders {
			slice[i] = copyOrder(o)
		}
		asksCopy[price] = slice
	}

	userOrderMapCopy := make(map[string]map[string]*Order, len(r.UserOrderMap))
	for userId, orders := range r.UserOrderMap {
		userOrderMapCopy[userId] = make(map[string]*Order, len(orders))
		for orderId, o := range orders {
			userOrderMapCopy[userId][orderId] = copyOrder(o)
		}
	}

	return OrderBook{
		Bids:         bidsCopy,
		Asks:         asksCopy,
		BidHeap:      r.BidHeap.Clone(),
		AskHeap:      r.AskHeap.Clone(),
		UserOrderMap: userOrderMapCopy,
		CurrentPrice: r.CurrentPrice,
		LastTradeId:  r.LastTradeId,
		LastOrderId:  r.LastOrderId,
		LastStreamId: r.LastStreamId,
	}
}

func (r *OrderBook) addOrderToBids(order *Order, price int) {
	isNewLevel := false
	if _, exists := r.Bids[price]; !exists {
		r.BidHeap.Insert(price)
		isNewLevel = true
	}
	r.Bids[price] = append(r.Bids[price], order)
	log.Printf("[BID QUEUED] orderId=%s userId=%s price=%d qty=%d filled=%d newLevel=%v",
		order.Id, order.UserId, price, order.Quantity, order.Filled, isNewLevel)

	if _, exists := r.UserOrderMap[order.UserId]; !exists {
		r.UserOrderMap[order.UserId] = make(map[string]*Order)
	}
	r.UserOrderMap[order.UserId][order.Id] = order
}

func (r *OrderBook) addOrderToAsks(order *Order, price int) {
	isNewLevel := false
	if _, ok := r.Asks[price]; !ok {
		r.AskHeap.Insert(price)
		isNewLevel = true
	}
	r.Asks[price] = append(r.Asks[price], order)
	log.Printf("[ASK QUEUED] orderId=%s userId=%s price=%d qty=%d filled=%d newLevel=%v",
		order.Id, order.UserId, price, order.Quantity, order.Filled, isNewLevel)

	if _, exists := r.UserOrderMap[order.UserId]; !exists {
		r.UserOrderMap[order.UserId] = make(map[string]*Order)
	}
	r.UserOrderMap[order.UserId][order.Id] = order
}

func (r *OrderBook) matchBids(order *Order, price int) ([]Fills, int) {
	var fills []Fills
	var executedQty int = 0
	var queued bool

	log.Printf("[MATCH BID START] orderId=%s userId=%s price=%d qty=%d askLevels=%d",
		order.Id, order.UserId, price, order.Quantity, r.AskHeap.Size())

	for order.Filled < order.Quantity {
		if r.AskHeap.Size() == 0 {
			log.Printf("[MATCH BID] no asks available — queuing orderId=%s at price=%d", order.Id, price)
			r.addOrderToBids(order, price)
			queued = true
			break
		}

		bestAskPrice := r.AskHeap.Peek()
		log.Printf("[MATCH BID] bestAsk=%d incomingBid=%d orderId=%s remaining=%d",
			bestAskPrice, price, order.Id, order.Quantity-order.Filled)

		if price < bestAskPrice {
			log.Printf("[MATCH BID] bid price %d < bestAsk %d — queuing orderId=%s", price, bestAskPrice, order.Id)
			r.addOrderToBids(order, price)
			queued = true
			break
		}

		askOrders := r.Asks[bestAskPrice]
		if len(askOrders) == 0 {
			log.Printf("[MATCH BID] empty ask level at price=%d — popping heap", bestAskPrice)
			r.AskHeap.Pop()
			continue
		}

		matchedInThisLevel := false

		for i := len(askOrders) - 1; i >= 0; i-- {
			if order.Filled >= order.Quantity {
				break
			}

			queuedOrder := askOrders[i]

			if queuedOrder.Filled >= queuedOrder.Quantity {
				log.Printf("[MATCH BID] stale ask orderId=%s at price=%d — removing", queuedOrder.Id, bestAskPrice)
				askOrders = append(askOrders[:i], askOrders[i+1:]...)
				continue
			}

			remainingInQueue := queuedOrder.Quantity - queuedOrder.Filled
			remainingInIncoming := order.Quantity - order.Filled

			matchQuantity := remainingInQueue
			if remainingInIncoming < remainingInQueue {
				matchQuantity = remainingInIncoming
			}

			order.Filled += matchQuantity
			queuedOrder.Filled += matchQuantity
			executedQty += matchQuantity
			matchedInThisLevel = true

			tradeId := uuid.NewString()
			fills = append(fills, Fills{
				Price:       bestAskPrice,
				Quantity:    matchQuantity,
				OtherUserId: queuedOrder.UserId,
				OrderId:     order.Id,
				TradeId:     tradeId,
			})

			log.Printf("[FILL] tradeId=%s price=%d qty=%d buyerId=%s sellerId=%s bidFilled=%d/%d askFilled=%d/%d",
				tradeId, bestAskPrice, matchQuantity,
				order.UserId, queuedOrder.UserId,
				order.Filled, order.Quantity,
				queuedOrder.Filled, queuedOrder.Quantity)

			r.CurrentPrice = bestAskPrice
			r.LastTradeId = tradeId

			if queuedOrder.Filled == queuedOrder.Quantity {
				log.Printf("[MATCH BID] ask fully filled orderId=%s — removing from level %d", queuedOrder.Id, bestAskPrice)
				askOrders = append(askOrders[:i], askOrders[i+1:]...)
			}
		}

		r.Asks[bestAskPrice] = askOrders

		if len(askOrders) == 0 {
			log.Printf("[MATCH BID] ask level %d exhausted — removing", bestAskPrice)
			delete(r.Asks, bestAskPrice)
			r.AskHeap.Pop()
		}

		if !matchedInThisLevel {
			break
		}
	}

	if !queued && order.Filled < order.Quantity {
		log.Printf("[MATCH BID] partial fill — queuing remainder orderId=%s filled=%d/%d at price=%d",
			order.Id, order.Filled, order.Quantity, price)
		r.addOrderToBids(order, price)
	} else if order.Filled >= order.Quantity {
		log.Printf("[MATCH BID COMPLETE] orderId=%s fully filled qty=%d executedQty=%d",
			order.Id, order.Quantity, executedQty)
	}

	return fills, executedQty
}

func (r *OrderBook) matchAsks(order *Order, price int) ([]Fills, int) {
	var fills []Fills
	var executedQty int = 0
	var queued bool

	log.Printf("[MATCH ASK START] orderId=%s userId=%s price=%d qty=%d bidLevels=%d",
		order.Id, order.UserId, price, order.Quantity, r.BidHeap.Size())

	for order.Filled < order.Quantity {
		if r.BidHeap.Size() == 0 {
			log.Printf("[MATCH ASK] no bids available — queuing orderId=%s at price=%d", order.Id, price)
			r.addOrderToAsks(order, price)
			queued = true
			break
		}

		bestBidPrice := r.BidHeap.Peek()
		log.Printf("[MATCH ASK] bestBid=%d incomingAsk=%d orderId=%s remaining=%d",
			bestBidPrice, price, order.Id, order.Quantity-order.Filled)

		if price > bestBidPrice {
			log.Printf("[MATCH ASK] ask price %d > bestBid %d — queuing orderId=%s", price, bestBidPrice, order.Id)
			r.addOrderToAsks(order, price)
			queued = true
			break
		}

		bidOrders := r.Bids[bestBidPrice]
		if len(bidOrders) == 0 {
			log.Printf("[MATCH ASK] empty bid level at price=%d — popping heap", bestBidPrice)
			r.BidHeap.Pop()
			continue
		}

		matchedInThisLevel := false

		for i := len(bidOrders) - 1; i >= 0; i-- {
			if order.Filled >= order.Quantity {
				break
			}

			queuedOrder := bidOrders[i]

			if queuedOrder.Filled >= queuedOrder.Quantity {
				log.Printf("[MATCH ASK] stale bid orderId=%s at price=%d — removing", queuedOrder.Id, bestBidPrice)
				bidOrders = append(bidOrders[:i], bidOrders[i+1:]...)
				continue
			}

			remainingInQueue := queuedOrder.Quantity - queuedOrder.Filled
			remainingInIncoming := order.Quantity - order.Filled

			matchQuantity := remainingInQueue
			if remainingInIncoming < remainingInQueue {
				matchQuantity = remainingInIncoming
			}

			order.Filled += matchQuantity
			queuedOrder.Filled += matchQuantity
			executedQty += matchQuantity
			matchedInThisLevel = true

			tradeId := uuid.NewString()
			fills = append(fills, Fills{
				Price:       bestBidPrice,
				Quantity:    matchQuantity,
				OtherUserId: queuedOrder.UserId,
				OrderId:     order.Id,
				TradeId:     tradeId,
			})

			log.Printf("[FILL] tradeId=%s price=%d qty=%d sellerId=%s buyerId=%s askFilled=%d/%d bidFilled=%d/%d",
				tradeId, bestBidPrice, matchQuantity,
				order.UserId, queuedOrder.UserId,
				order.Filled, order.Quantity,
				queuedOrder.Filled, queuedOrder.Quantity)

			r.CurrentPrice = bestBidPrice
			r.LastTradeId = tradeId

			if queuedOrder.Filled == queuedOrder.Quantity {
				log.Printf("[MATCH ASK] bid fully filled orderId=%s — removing from level %d", queuedOrder.Id, bestBidPrice)
				bidOrders = append(bidOrders[:i], bidOrders[i+1:]...)
			}
		}

		r.Bids[bestBidPrice] = bidOrders

		if len(bidOrders) == 0 {
			log.Printf("[MATCH ASK] bid level %d exhausted — removing", bestBidPrice)
			delete(r.Bids, bestBidPrice)
			r.BidHeap.Pop()
		}

		if !matchedInThisLevel {
			break
		}
	}

	if !queued && order.Filled < order.Quantity {
		log.Printf("[MATCH ASK] partial fill — queuing remainder orderId=%s filled=%d/%d at price=%d",
			order.Id, order.Filled, order.Quantity, price)
		r.addOrderToAsks(order, price)
	} else if order.Filled >= order.Quantity {
		log.Printf("[MATCH ASK COMPLETE] orderId=%s fully filled qty=%d executedQty=%d",
			order.Id, order.Quantity, executedQty)
	}

	return fills, executedQty
}

func (r *OrderBook) GetDepth() (map[int][]*Order, map[int][]*Order) {
	return r.Bids, r.Asks
}

func (r *OrderBook) AddOrder(order Order, price int) ([]Fills, int, error) {
	slog.Info("ORDER RECIEVED ::", order)

	r.LastOrderId = order.Id
	if order.Side == BUY {
		fills, executedQty := r.matchBids(&order, price)
		r.LastStreamId = order.StreamId
		return fills, executedQty, nil
	} else if order.Side == SELL {
		fills, executedQty := r.matchAsks(&order, price)
		r.LastStreamId = order.StreamId
		return fills, executedQty, nil
	}

	return []Fills{}, 0, errors.New("invalid order side")
}

func (r *OrderBook) GetOpenOrders(userId string) []*Order {
	var result []*Order

	userOrders, ok := r.UserOrderMap[userId]
	if !ok {
		return result
	}

	for _, order := range userOrders {
		if order.Filled < order.Quantity {
			result = append(result, order)
		}
	}

	return result
}

func (r *OrderBook) CancelOrder(orderId string, userId string) bool {
	userOrders, ok := r.UserOrderMap[userId]
	if !ok {
		log.Printf("[CANCEL] userId=%s not found in UserOrderMap", userId)
		return false
	}

	_, exists := userOrders[orderId]
	if !exists {
		log.Printf("[CANCEL] orderId=%s not found for userId=%s", orderId, userId)
		return false
	}

	for price, orders := range r.Bids {
		for i, o := range orders {
			if o.Id == orderId {
				r.Bids[price] = append(r.Bids[price][:i], r.Bids[price][i+1:]...)
				if len(r.Bids[price]) == 0 {
					delete(r.Bids, price)
				}
				delete(userOrders, orderId)
				log.Printf("[CANCEL] BID cancelled orderId=%s userId=%s price=%d filled=%d/%d",
					orderId, userId, price, o.Filled, o.Quantity)
				return true
			}
		}
	}

	for price, orders := range r.Asks {
		for i, o := range orders {
			if o.Id == orderId {
				r.Asks[price] = append(r.Asks[price][:i], r.Asks[price][i+1:]...)
				if len(r.Asks[price]) == 0 {
					delete(r.Asks, price)
				}
				delete(userOrders, orderId)
				log.Printf("[CANCEL] ASK cancelled orderId=%s userId=%s price=%d filled=%d/%d",
					orderId, userId, price, o.Filled, o.Quantity)
				return true
			}
		}
	}

	log.Printf("[CANCEL] orderId=%s userId=%s found in UserOrderMap but not in Bids/Asks", orderId, userId)
	return false
}

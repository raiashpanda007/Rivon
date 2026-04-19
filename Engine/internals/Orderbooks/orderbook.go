package orderbooks

import (
	"errors"

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
	Price        int
	Quantity     int
	TradeId      string
	OtherUserId  string
	OtherOrderId string
	OrderId      string
}

type OrderBook struct {
	Bids         map[int][]*Order
	Asks         map[int][]*Order
	BidHeap      *heap.MaxHeap
	AskHeap      *heap.MinHeap
	BidDepth     map[int]int
	AskDepth     map[int]int
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
	bidDepth := make(map[int]int)
	askDepth := make(map[int]int)

	for price, orders := range bids {
		for i := range orders {
			orderPtr := &orders[i]
			bidsCopy[price] = append(bidsCopy[price], orderPtr)
			bidDepth[price] += orderPtr.Quantity - orderPtr.Filled

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
			askDepth[price] += orderPtr.Quantity - orderPtr.Filled

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
		BidDepth:     bidDepth,
		AskDepth:     askDepth,
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

	bidDepthCopy := make(map[int]int, len(r.BidDepth))
	for price, qty := range r.BidDepth {
		bidDepthCopy[price] = qty
	}
	askDepthCopy := make(map[int]int, len(r.AskDepth))
	for price, qty := range r.AskDepth {
		askDepthCopy[price] = qty
	}

	return OrderBook{
		Bids:         bidsCopy,
		Asks:         asksCopy,
		BidHeap:      r.BidHeap.Clone(),
		AskHeap:      r.AskHeap.Clone(),
		BidDepth:     bidDepthCopy,
		AskDepth:     askDepthCopy,
		UserOrderMap: userOrderMapCopy,
		CurrentPrice: r.CurrentPrice,
		LastTradeId:  r.LastTradeId,
		LastOrderId:  r.LastOrderId,
		LastStreamId: r.LastStreamId,
	}
}

func (r *OrderBook) addOrderToBids(order *Order, price int) {
	if _, exists := r.Bids[price]; !exists {
		r.BidHeap.Insert(price)
	}
	r.Bids[price] = append(r.Bids[price], order)
	r.BidDepth[price] += order.Quantity - order.Filled

	if _, exists := r.UserOrderMap[order.UserId]; !exists {
		r.UserOrderMap[order.UserId] = make(map[string]*Order)
	}
	r.UserOrderMap[order.UserId][order.Id] = order
}

func (r *OrderBook) addOrderToAsks(order *Order, price int) {
	if _, ok := r.Asks[price]; !ok {
		r.AskHeap.Insert(price)
	}
	r.Asks[price] = append(r.Asks[price], order)
	r.AskDepth[price] += order.Quantity - order.Filled

	if _, exists := r.UserOrderMap[order.UserId]; !exists {
		r.UserOrderMap[order.UserId] = make(map[string]*Order)
	}
	r.UserOrderMap[order.UserId][order.Id] = order
}

func (r *OrderBook) matchBids(order *Order, price int) ([]Fills, int) {
	var fills []Fills
	var executedQty int = 0
	var queued bool

	for order.Filled < order.Quantity {
		if r.AskHeap.Size() == 0 {
			r.addOrderToBids(order, price)
			queued = true
			break
		}

		bestAskPrice := r.AskHeap.Peek()

		if price < bestAskPrice {
			r.addOrderToBids(order, price)
			queued = true
			break
		}

		askOrders := r.Asks[bestAskPrice]
		if len(askOrders) == 0 {
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
			r.AskDepth[bestAskPrice] -= matchQuantity

			tradeId := uuid.NewString()
			fills = append(fills, Fills{
				Price:        bestAskPrice,
				Quantity:     matchQuantity,
				OtherUserId:  queuedOrder.UserId,
				OtherOrderId: queuedOrder.Id,
				OrderId:      order.Id,
				TradeId:      tradeId,
			})

			r.CurrentPrice = bestAskPrice
			r.LastTradeId = tradeId

			if queuedOrder.Filled == queuedOrder.Quantity {
				askOrders = append(askOrders[:i], askOrders[i+1:]...)
			}
		}

		r.Asks[bestAskPrice] = askOrders

		if len(askOrders) == 0 {
			delete(r.Asks, bestAskPrice)
			delete(r.AskDepth, bestAskPrice)
			r.AskHeap.Pop()
		}

		if !matchedInThisLevel {
			break
		}
	}

	if !queued && order.Filled < order.Quantity {
		r.addOrderToBids(order, price)
	}

	return fills, executedQty
}

func (r *OrderBook) matchAsks(order *Order, price int) ([]Fills, int) {
	var fills []Fills
	var executedQty int = 0
	var queued bool

	for order.Filled < order.Quantity {
		if r.BidHeap.Size() == 0 {
			r.addOrderToAsks(order, price)
			queued = true
			break
		}

		bestBidPrice := r.BidHeap.Peek()

		if price > bestBidPrice {
			r.addOrderToAsks(order, price)
			queued = true
			break
		}

		bidOrders := r.Bids[bestBidPrice]
		if len(bidOrders) == 0 {
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
			r.BidDepth[bestBidPrice] -= matchQuantity

			tradeId := uuid.NewString()
			fills = append(fills, Fills{
				Price:        bestBidPrice,
				Quantity:     matchQuantity,
				OtherUserId:  queuedOrder.UserId,
				OtherOrderId: queuedOrder.Id,
				OrderId:      order.Id,
				TradeId:      tradeId,
			})

			r.CurrentPrice = bestBidPrice
			r.LastTradeId = tradeId

			if queuedOrder.Filled == queuedOrder.Quantity {
				bidOrders = append(bidOrders[:i], bidOrders[i+1:]...)
			}
		}

		r.Bids[bestBidPrice] = bidOrders

		if len(bidOrders) == 0 {
			delete(r.Bids, bestBidPrice)
			delete(r.BidDepth, bestBidPrice)
			r.BidHeap.Pop()
		}

		if !matchedInThisLevel {
			break
		}
	}

	if !queued && order.Filled < order.Quantity {
		r.addOrderToAsks(order, price)
	}

	return fills, executedQty
}

func (r *OrderBook) AddOrder(order Order, price int) ([]Fills, int, error) {
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

func (r *OrderBook) CancelOrder(orderId string, userId string) (*Order, bool) {
	userOrders, ok := r.UserOrderMap[userId]
	if !ok {
		return nil, false
	}

	order, exists := userOrders[orderId]
	if !exists {
		return nil, false
	}

	var bucket map[int][]*Order
	if order.Side == BUY {
		bucket = r.Bids
	} else {
		bucket = r.Asks
	}

	var depthMap map[int]int
	if order.Side == BUY {
		depthMap = r.BidDepth
	} else {
		depthMap = r.AskDepth
	}

	orders := bucket[order.Price]
	for i, o := range orders {
		if o.Id == orderId {
			bucket[order.Price] = append(orders[:i], orders[i+1:]...)
			depthMap[order.Price] -= order.Quantity - order.Filled
			if len(bucket[order.Price]) == 0 {
				delete(bucket, order.Price)
				delete(depthMap, order.Price)
				if order.Side == BUY {
					r.BidHeap.Remove(order.Price)
				} else {
					r.AskHeap.Remove(order.Price)
				}
			}
			delete(userOrders, orderId)
			return order, true
		}
	}

	return nil, false
}

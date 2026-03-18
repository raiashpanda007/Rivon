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
}

func NewOrderBook(lastOrderId string, bids map[int][]Order, asks map[int][]Order, askHeap *heap.MinHeap, bidHeap *heap.MaxHeap, currentPrice int) OrderBook {
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
		LastTradeId:  lastOrderId,
	}
}

func (r *OrderBook) GetSnapshot() *OrderBook {
	return r
}

func (r *OrderBook) addOrderToBids(order *Order, price int) {
	if _, exists := r.Bids[price]; !exists {
		r.BidHeap.Insert(price)
	}
	r.Bids[price] = append(r.Bids[price], order)

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

	if _, exists := r.UserOrderMap[order.UserId]; !exists {
		r.UserOrderMap[order.UserId] = make(map[string]*Order)
	}
	r.UserOrderMap[order.UserId][order.Id] = order
}

func (r *OrderBook) matchBids(order *Order, price int) ([]Fills, int) {
	var fills []Fills
	var executedQty int = 0

	for order.Filled < order.Quantity {
		if r.AskHeap.Size() == 0 {
			r.addOrderToBids(order, price)
			break
		}

		bestAskPrice := r.AskHeap.Peek()

		if price < bestAskPrice {
			r.addOrderToBids(order, price)
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

			tradeId := uuid.NewString()
			fills = append(fills, Fills{
				Price:       bestAskPrice,
				Quantity:    matchQuantity,
				OtherUserId: queuedOrder.UserId,
				OrderId:     order.Id,
				TradeId:     tradeId,
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
			r.AskHeap.Pop()
		}

		if !matchedInThisLevel {
			break
		}
	}

	if order.Filled < order.Quantity {
		r.addOrderToBids(order, price)
	}

	return fills, executedQty
}

func (r *OrderBook) matchAsks(order *Order, price int) ([]Fills, int) {
	var fills []Fills
	var executedQty int = 0

	for order.Filled < order.Quantity {
		if r.BidHeap.Size() == 0 {
			r.addOrderToAsks(order, price)
			break
		}

		bestBidPrice := r.BidHeap.Peek()

		if price > bestBidPrice {
			r.addOrderToAsks(order, price)
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

			tradeId := uuid.NewString()
			fills = append(fills, Fills{
				Price:       bestBidPrice,
				Quantity:    matchQuantity,
				OtherUserId: queuedOrder.UserId,
				OrderId:     order.Id,
				TradeId:     tradeId,
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
			r.BidHeap.Pop()
		}

		if !matchedInThisLevel {
			break
		}
	}

	if order.Filled < order.Quantity {
		r.addOrderToAsks(order, price)
	}

	return fills, executedQty
}

func (r *OrderBook) GetDepth() (map[int][]*Order, map[int][]*Order) {
	return r.Bids, r.Asks
}

func (r *OrderBook) AddOrder(order Order, price int) ([]Fills, int, error) {
	if order.Side == BUY {
		fills, executedQty := r.matchBids(&order, price)
		return fills, executedQty, nil
	} else if order.Side == SELL {
		fills, executedQty := r.matchAsks(&order, price)
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
		return false
	}

	_, exists := userOrders[orderId]
	if !exists {
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
				return true
			}
		}
	}

	return false
}

package orderbooks

import heap "github.com/raiashpanda007/rivon/engine/internals/utils"

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
}

type Fills struct {
	Price    int
	Quantity int
	TradeId  string
	UserId   string
	OrderId  string
}

type OrderBook struct {
	Bids         map[int][]*Order
	Asks         map[int][]*Order
	BidHeap      *heap.MaxHeap
	AskHeap      *heap.MinHeap
	UserOrderMap map[string]map[string]*Order
	CurrentPrice int64
	LastTradeId  string
}

func NewOrderBook(lastOrderId string, bids map[int][]Order, asks map[int][]Order, askHeap *heap.MinHeap, bidHeap *heap.MaxHeap, currentPrice int64) OrderBook {
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

func (r *OrderBook) AddOrder(order Order, price int) ([]Fills, int) {

	return []Fills{}, 0
}

func (r *OrderBook) matchBids(order Order, price int) ([]Fills, int) {
	return []Fills{}, 0
}

func (r *OrderBook) matchAsks(order Order, price int) ([]Fills, int) {
	return []Fills{}, 0
}
func (r *OrderBook) GetDepth() (map[int][]Order, map[int][]Order) {
	return map[int][]Order{}, map[int][]Order{}
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

func (r *OrderBook) CancelOrder(order Order) bool {
	return true
}

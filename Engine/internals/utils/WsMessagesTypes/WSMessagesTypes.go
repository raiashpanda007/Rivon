package wsmessagestypes

type WSInMessageType string

type WSOutMessageType string

type WSOutPayload interface {
	wsPaylod()
}

const (
	ORDERBOOK_SUBSCIRBE WSInMessageType = "ORDER_BOOK_SUBSCRIBE"
	DEPTH_SUBSCRIBE     WSInMessageType = "DEPTH_SUBSCRIBE"
	WALLET_LOAD         WSInMessageType = "WALLET_LOAD"
	WALLET_EVICT        WSInMessageType = "WALLET_EVICT"
	CANCEL_ORDER_WS     WSInMessageType = "CANCEL_ORDER_WS"
)

const (
	ORDERBOOK_DATA   WSOutMessageType = "ORDERBOOK_DATA"
	DEPTH_DATA       WSOutMessageType = "DEPTH_DATA"
	ORDERBOOK_UPDATE WSOutMessageType = "ORDERBOOK_UPDATE"
	ORDER_CANCELLED  WSOutMessageType = "ORDER_CANCELLED"
)

type WSOutMessageStruct struct {
	MessageType  WSOutMessageType `json:"type"`
	Payload      WSOutPayload     `json:"payload"`
	UserId       string           `json:"userId,omitempty"`
	ConnectionId string           `json:"connectionId"`
}

// WSInMessageStruct carries an inbound WS message.
// UserId is empty for unauthenticated connections; ConnectionId always
// identifies the physical connection so the WS server can route the reply.
type WSInMessageStruct struct {
	MessageType  WSInMessageType `json:"MessageType"`
	UserId       string          `json:"UserId"`
	ConnectionId string          `json:"ConnectionId"`
	OrderId      string          `json:"OrderId,omitempty"`
	CancelQty    int             `json:"CancelQty,omitempty"`
}

type OrderbookPayload struct {
	BidDepth     map[int]int `json:"bidDepth"`
	AskDepth     map[int]int `json:"askDepth"`
	CurrentPrice int         `json:"currentPrice"`
}

func (o OrderbookPayload) wsPaylod() {}

type DepthPayload struct {
	BidDepth map[int]int `json:"bidDepth"`
	AskDepth map[int]int `json:"askDepth"`
}

func (d DepthPayload) wsPaylod() {}

// PublicFill is a fill without user identifiers — safe to broadcast.
type PublicFill struct {
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	TradeId  string `json:"tradeId"`
}

type OrderbookUpdatePayload struct {
	BidDepth     map[int]int  `json:"bidDepth"`
	AskDepth     map[int]int  `json:"askDepth"`
	CurrentPrice int          `json:"currentPrice"`
	Fills        []PublicFill `json:"fills"`
}

func (o OrderbookUpdatePayload) wsPaylod() {}

type OrderCancelledPayload struct {
	OrderId      string `json:"orderId"`
	Success      bool   `json:"success"`
	CancelledQty int    `json:"cancelledQty"`
}

func (o OrderCancelledPayload) wsPaylod() {}

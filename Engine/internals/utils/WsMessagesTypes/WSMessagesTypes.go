package wsmessagestypes

type WSInMessageType string

type WSOutMessageType string

type WSOutPayload interface {
	wsPaylod()
}

const (
	ORDERBOOK_SUBSCIRBE WSInMessageType = "ORDER_BOOK_SUBSCRIBE"
	DEPTH_SUBSCRIBE     WSInMessageType = "DEPTH_SUBSCRIBE"
)

const (
	ORDERBOOK_DATA   WSOutMessageType = "ORDERBOOK_DATA"
	DEPTH_DATA       WSOutMessageType = "DEPTH_DATA"
	ORDERBOOK_UPDATE WSOutMessageType = "ORDERBOOK_UPDATE"
)

type WSOutMessageStruct struct {
	MessageType WSOutMessageType `json:"type"`
	Payload     WSOutPayload     `json:"payload"`
	UserId      string           `json:"userId"`
}

type WSInMessageStruct struct {
	MessageType WSInMessageType
	UserId      string
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

package markets

import (
	"log/slog"
)

type OrderMessages struct {
	OrderId   string `json:"orderId"`
	UserId    string `json:"userId"`
	MarketId  string `json:"marketId"`
	Price     int    `json:"price"`
	Quantity  int    `json:"quantity"`
	OrderType string `json:"orderType"`
}

func StarMarketProcess(ch chan OrderMessages) {

	for order := range ch {
		slog.Info("Order recieved here :: ", order)
	}
}

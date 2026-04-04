package snapshots

import (
	"fmt"
	"log"
	"os"
	"time"

	orderbooks "github.com/raiashpanda007/rivon/engine/internals/Orderbooks"
	"github.com/vmihailenco/msgpack/v5"
)

func SaveSnapShot(marketId string, orderbook orderbooks.OrderBook) {
	snapshot := orderbook.GetSnapshot()

	data, err := msgpack.Marshal(snapshot)
	if err != nil {
		log.Printf("Snapshot marshal error for market %s: %v", marketId, err)
		return
	}

	dir := "../snapshots"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Snapshot dir create error: %v", err)
		return
	}

	fileName := fmt.Sprintf("%s/%v.%d.mpac", dir, marketId, time.Now().Unix())
	if err := os.WriteFile(fileName, data, 0644); err != nil {
		log.Printf("Snapshot write error for market %s: %v", marketId, err)
		return
	}

	log.Printf("Snapshot saved: %s", fileName)
}

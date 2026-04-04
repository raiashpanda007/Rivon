package snapshots

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	orderbooks "github.com/raiashpanda007/rivon/engine/internals/Orderbooks"
	heap "github.com/raiashpanda007/rivon/engine/internals/utils"
	"github.com/vmihailenco/msgpack/v5"
)

// snapshotData is the on-disk format for an OrderBook.
// It uses a flat order list + price→IDs maps to avoid pointer-sharing problems
// that arise when serialising map[int][]*Order directly.
type snapshotData struct {
	Orders       []orderbooks.Order
	BidOrderIds  map[int][]string
	AskOrderIds  map[int][]string
	CurrentPrice int
	LastTradeId  string
	LastOrderId  string
	LastStreamId string
}

// toSnapshotData converts a live OrderBook into the serialisable form.
func toSnapshotData(ob orderbooks.OrderBook) snapshotData {
	seen := make(map[string]bool)
	var orders []orderbooks.Order

	bidOrderIds := make(map[int][]string, len(ob.Bids))
	for price, priceOrders := range ob.Bids {
		ids := make([]string, 0, len(priceOrders))
		for _, o := range priceOrders {
			ids = append(ids, o.Id)
			if !seen[o.Id] {
				seen[o.Id] = true
				orders = append(orders, *o)
			}
		}
		bidOrderIds[price] = ids
	}

	askOrderIds := make(map[int][]string, len(ob.Asks))
	for price, priceOrders := range ob.Asks {
		ids := make([]string, 0, len(priceOrders))
		for _, o := range priceOrders {
			ids = append(ids, o.Id)
			if !seen[o.Id] {
				seen[o.Id] = true
				orders = append(orders, *o)
			}
		}
		askOrderIds[price] = ids
	}

	return snapshotData{
		Orders:       orders,
		BidOrderIds:  bidOrderIds,
		AskOrderIds:  askOrderIds,
		CurrentPrice: ob.CurrentPrice,
		LastTradeId:  ob.LastTradeId,
		LastOrderId:  ob.LastOrderId,
		LastStreamId: ob.LastStreamId,
	}
}

// fromSnapshotData reconstructs an OrderBook from the serialised form.
// It uses NewOrderBook so that pointer-sharing between Bids/Asks/UserOrderMap
// is correctly established.
func fromSnapshotData(sd snapshotData) orderbooks.OrderBook {
	orderById := make(map[string]orderbooks.Order, len(sd.Orders))
	for _, o := range sd.Orders {
		orderById[o.Id] = o
	}

	bidHeap := heap.NewMaxHeap()
	bids := make(map[int][]orderbooks.Order, len(sd.BidOrderIds))
	for price, ids := range sd.BidOrderIds {
		orders := make([]orderbooks.Order, 0, len(ids))
		for _, id := range ids {
			if o, ok := orderById[id]; ok {
				orders = append(orders, o)
			}
		}
		if len(orders) > 0 {
			bids[price] = orders
			bidHeap.Insert(price)
		}
	}

	askHeap := heap.NewMinHeap()
	asks := make(map[int][]orderbooks.Order, len(sd.AskOrderIds))
	for price, ids := range sd.AskOrderIds {
		orders := make([]orderbooks.Order, 0, len(ids))
		for _, id := range ids {
			if o, ok := orderById[id]; ok {
				orders = append(orders, o)
			}
		}
		if len(orders) > 0 {
			asks[price] = orders
			askHeap.Insert(price)
		}
	}

	return orderbooks.NewOrderBook(
		sd.LastTradeId,
		sd.LastOrderId,
		sd.LastStreamId,
		bids,
		asks,
		askHeap,
		bidHeap,
		sd.CurrentPrice,
	)
}

const maxSnapshotsToKeep = 3

// SaveSnapShot serialises ob (which should already be an independent snapshot copy)
// to a msgpack file under ../snapshots/<marketId>/ and prunes old files.
func SaveSnapShot(marketId string, ob orderbooks.OrderBook) {
	dir := fmt.Sprintf("../snapshots/%s", marketId)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Snapshot dir create error: %v", err)
		return
	}

	sd := toSnapshotData(ob)

	data, err := msgpack.Marshal(sd)
	if err != nil {
		log.Printf("Snapshot marshal error for market %s: %v", marketId, err)
		return
	}

	fileName := fmt.Sprintf("%s/%d.mpac", dir, time.Now().Unix())
	if err := os.WriteFile(fileName, data, 0644); err != nil {
		log.Printf("Snapshot write error for market %s: %v", marketId, err)
		return
	}

	log.Printf("Snapshot saved: %s (orders=%d bids=%d asks=%d)",
		fileName, len(sd.Orders), len(sd.BidOrderIds), len(sd.AskOrderIds))

	pruneSnapshots(dir)
}

// pruneSnapshots removes all but the most recent maxSnapshotsToKeep files.
func pruneSnapshots(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) <= maxSnapshotsToKeep {
		return
	}

	type fileInfo struct {
		name    string
		modTime time.Time
	}

	var files []fileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{e.Name(), info.ModTime()})
	}

	// oldest first
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	for i := 0; i < len(files)-maxSnapshotsToKeep; i++ {
		path := fmt.Sprintf("%s/%s", dir, files[i].name)
		if err := os.Remove(path); err != nil {
			log.Printf("Snapshot prune error: %v", err)
		} else {
			log.Printf("Snapshot pruned: %s", path)
		}
	}
}

// ReadLastSnapShotForMarket reads the most recent snapshot for a market and
// returns a fully reconstructed OrderBook. Returns (nil, false) if no snapshot exists.
func ReadLastSnapShotForMarket(marketId string) (*orderbooks.OrderBook, bool) {
	parentDir := fmt.Sprintf("../snapshots/%s", marketId)

	entries, err := os.ReadDir(parentDir)
	if err != nil {
		log.Printf("No snapshot directory for marketId %s: %v", marketId, err)
		return nil, false
	}

	var latestFile os.DirEntry
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestFile = entry
		}
	}

	if latestFile == nil {
		log.Printf("No snapshot files for marketId %s", marketId)
		return nil, false
	}

	snapshotPath := fmt.Sprintf("%s/%s", parentDir, latestFile.Name())

	fileData, err := os.ReadFile(snapshotPath)
	if err != nil {
		log.Printf("Unable to read snapshot %s: %v", snapshotPath, err)
		return nil, false
	}

	var sd snapshotData
	if err := msgpack.Unmarshal(fileData, &sd); err != nil {
		log.Printf("Invalid msgpack in %s: %v", snapshotPath, err)
		return nil, false
	}

	ob := fromSnapshotData(sd)
	log.Printf("Snapshot loaded: %s (orders=%d bids=%d asks=%d currentPrice=%d lastStreamId=%s)",
		snapshotPath, len(sd.Orders), len(sd.BidOrderIds), len(sd.AskOrderIds),
		sd.CurrentPrice, sd.LastStreamId)

	return &ob, true
}

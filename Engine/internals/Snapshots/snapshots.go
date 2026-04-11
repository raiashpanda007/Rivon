package snapshots

import (
	"fmt"
	"os"
	"sort"
	"time"

	orderbooks "github.com/raiashpanda007/rivon/engine/internals/Orderbooks"
	heap "github.com/raiashpanda007/rivon/engine/internals/utils"
	"github.com/vmihailenco/msgpack/v5"
)

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
		return
	}

	sd := toSnapshotData(ob)

	data, err := msgpack.Marshal(sd)
	if err != nil {
		return
	}

	fileName := fmt.Sprintf("%s/%d.mpac", dir, time.Now().Unix())
	if err := os.WriteFile(fileName, data, 0644); err != nil {
		return
	}

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
		os.Remove(path)
	}
}

// ReadLastSnapShotForMarket reads the most recent snapshot for a market and
// returns a fully reconstructed OrderBook. Returns (nil, false) if no snapshot exists.
func ReadLastSnapShotForMarket(marketId string) (*orderbooks.OrderBook, bool) {
	parentDir := fmt.Sprintf("../snapshots/%s", marketId)

	entries, err := os.ReadDir(parentDir)
	if err != nil {
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
		return nil, false
	}

	snapshotPath := fmt.Sprintf("%s/%s", parentDir, latestFile.Name())

	fileData, err := os.ReadFile(snapshotPath)
	if err != nil {
		return nil, false
	}

	var sd snapshotData
	if err := msgpack.Unmarshal(fileData, &sd); err != nil {
		return nil, false
	}

	ob := fromSnapshotData(sd)
	return &ob, true
}

package registry

import (
	"sync"

	"github.com/raiashpanda007/rivon/internals/types"
)

type Registry struct {
	mu      sync.Mutex
	pending map[string]chan types.FillResult
}

func New() *Registry {
	return &Registry{
		pending: make(map[string]chan types.FillResult),
	}
}

// Register creates a buffered channel for the given orderId and stores it.
// Buffer size 1 ensures Resolve never blocks even if PlaceOrder has already timed out.
func (r *Registry) Register(orderId string) chan types.FillResult {
	ch := make(chan types.FillResult, 1)
	r.mu.Lock()
	r.pending[orderId] = ch
	r.mu.Unlock()
	return ch
}

// Resolve sends the fill result to the waiting PlaceOrder call and removes the entry.
// It is a no-op if the orderId is not present (already timed out and deleted).
func (r *Registry) Resolve(orderId string, fill types.FillResult) {
	r.mu.Lock()
	ch, ok := r.pending[orderId]
	if ok {
		delete(r.pending, orderId)
	}
	r.mu.Unlock()
	if ok {
		ch <- fill
	}
}

// Delete removes a pending entry without sending. Called on timeout or context cancel.
func (r *Registry) Delete(orderId string) {
	r.mu.Lock()
	delete(r.pending, orderId)
	r.mu.Unlock()
}

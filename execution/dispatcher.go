package execution

import (
	"github.com/ethereum/go-ethereum/core/types"
	"sparseth/internal/log"
	"sync"
)

// Dispatcher manages subscriptions of new
// block headers and broadcasts them to
// multiple subscribers.
type Dispatcher struct {
	subs map[string]chan *types.Header
	log  log.Logger
	mu   sync.Mutex
}

// NewDispatcher returns a new dispatcher with
// the specified logger and no subscriptions.
func NewDispatcher(log log.Logger) *Dispatcher {
	return &Dispatcher{
		subs: make(map[string]chan *types.Header),
		log:  log.With("component", "dispatcher"),
	}
}

// Close closes and removes all
// subscriber channels.
func (d *Dispatcher) Close() {
	d.log.Info("shutting down")

	d.mu.Lock()
	defer d.mu.Unlock()

	for _, ch := range d.subs {
		close(ch)
	}

	d.subs = make(map[string]chan *types.Header)
}

// Subscribe registers a new subscriber to receive
// block headers. By default, a buffered channel is
// created. If the specified id is already subscribed,
// the existing channel is returned.
func (d *Dispatcher) Subscribe(id string) <-chan *types.Header {
	d.mu.Lock()
	defer d.mu.Unlock()

	if ch, exists := d.subs[id]; exists {
		return ch
	}

	d.log.Info("new subscription", "id", id)
	ch := make(chan *types.Header, 10)
	d.subs[id] = ch
	return ch
}

// Unsubscribe removes the subscriber with the
// given id and closes its channel. If no
// subscriber with the specified id exists,
// Unsubscribe does nothing.
func (d *Dispatcher) Unsubscribe(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if ch, exists := d.subs[id]; exists {
		d.log.Info("unsubscribe", "id", id)
		delete(d.subs, id)
		close(ch)
	}
}

// Broadcast sends the specified block header
// to all active subscribers.
func (d *Dispatcher) Broadcast(head *types.Header) {
	d.log.Info("received new block head", "hash", head.Hash())

	d.mu.Lock()
	defer d.mu.Unlock()

	for id, ch := range d.subs {
		select {
		case ch <- head:
		default:
			d.log.Warn("dropping block head for subscriber", "id", id, "head", head.Hash())
		}
	}
}

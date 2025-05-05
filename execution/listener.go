package execution

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"sparseth/internal/log"
)

// Listener subscribes to new block headers
// and dispatches them as they arrive.
type Listener struct {
	sub        <-chan *types.Header
	dispatcher *Dispatcher
	log        log.Logger
}

// NewListener creates a new block Listener that
// listens for block headers from the specified
// channel.
func NewListener(ch <-chan *types.Header, dispatcher *Dispatcher, log log.Logger) *Listener {
	return &Listener{
		sub:        ch,
		dispatcher: dispatcher,
		log:        log.With("component", "block-listener"),
	}
}

// RunContext starts listening for new block
// headers and processes them as they arrive.
func (l *Listener) RunContext(ctx context.Context) error {
	l.log.Info("start listening for block headers")

	for {
		select {
		case head := <-l.sub:
			l.log.Info("received new block head", "hash", head.Hash())
			l.dispatcher.Broadcast(head)
		case <-ctx.Done():
			l.log.Info("stop listening for block headers")
			return nil
		}
	}
}

package monitor

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"sparseth/log"
)

type Monitor struct {
	log log.Logger
	// sub is the channel for receiving
	// new block headers.
	sub <-chan *types.Header
	// processor handles business logic
	// to process blocks
	processor Processor
}

// NewMonitor creates a new Monitor for the
// specified Ethereum smart contract.
func NewMonitor(name string, ch <-chan *types.Header, processor Processor, log log.Logger) *Monitor {
	return &Monitor{
		log:       log.With("component", name+"-monitor"),
		sub:       ch,
		processor: processor,
	}
}

// RunContext starts the monitoring loop
// until the context is canceled.
func (m *Monitor) RunContext(ctx context.Context) error {
	m.log.Info("start monitor")

	for {
		select {
		case head := <-m.sub:
			if err := m.processBlock(ctx, head); err != nil {
				m.log.Warn("failed to process block", "num", head.Number, "hash", head.Hash().Hex(), "err", err)
			}
		case <-ctx.Done():
			m.log.Info("stop monitor")
			return nil
		}
	}
}

// processBlock handles a single block.
func (m *Monitor) processBlock(ctx context.Context, header *types.Header) error {
	m.log.Debug("process block", "num", header.Number, "hash", header.Hash().Hex())

	if err := m.processor.ProcessBlock(ctx, header); err != nil {
		return fmt.Errorf("failed to process block: %w", err)
	}

	m.log.Info("block verified", "num", header.Number, "hash", header.Hash().Hex())
	return nil
}

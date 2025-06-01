package monitor

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
)

// Processor defines the core interface
// for processing Ethereum block headers.
type Processor interface {
	// ProcessBlock handles a single block header.
	ProcessBlock(ctx context.Context, head *types.Header) error
}

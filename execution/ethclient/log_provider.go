package ethclient

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// logProvider retrieves logs for
// Ethereum accounts.
type logProvider struct {
	c *Client
}

// newLogProvider creates a new logProvider
// using the specified client.
func newLogProvider(client *Client) *logProvider {
	return &logProvider{
		c: client,
	}
}

// getLogsAtBlock retrieves logs for the specified
// Ethereum account at the specified block.
func (r *logProvider) getLogsAtBlock(ctx context.Context, account common.Address, blockNum *big.Int) ([]*types.Log, error) {
	return r.c.GetLogsAtBlock(ctx, account, blockNum)
}

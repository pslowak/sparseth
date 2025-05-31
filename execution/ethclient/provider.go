package ethclient

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// Provider is a wrapper for the Ethereum RPC API.
// Unlike Client, Provider may add additional
// verification to the provided data.
type Provider struct {
	log     *logProvider
	storage *storageProvider
}

// NewProvider creates a new Provider that
// uses the specified Ethereum RPC client.
func NewProvider(rpc *Client) *Provider {
	return &Provider{
		log:     newLogProvider(rpc),
		storage: newStorageProvider(rpc),
	}
}

// GetLogsAtBlock retrieves the logs for the specified
// Ethereum account at the specified block.
func (f *Provider) GetLogsAtBlock(ctx context.Context, acc common.Address, blockNum *big.Int) ([]*types.Log, error) {
	return f.log.getLogsAtBlock(ctx, acc, blockNum)
}

// GetStorageAtBlock provides the verified value stored at
// the specified storage slot for the specified Ethereum
// account at the specified block.
func (f *Provider) GetStorageAtBlock(ctx context.Context, acc common.Address, slot common.Hash, head *types.Header) ([]byte, error) {
	return f.storage.getSlot(ctx, acc, slot, head)
}

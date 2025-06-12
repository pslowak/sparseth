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
	tx  *txProvider
	log *logProvider
	acc *accountProvider
}

// NewProvider creates a new Provider that
// uses the specified Ethereum RPC client.
func NewProvider(rpc *Client) *Provider {
	return &Provider{
		tx:  newTxProvider(rpc),
		log: newLogProvider(rpc),
		acc: newAccountProvider(rpc),
	}
}

// GetTxsAtBlock retrieves all transactions at the
// specified block. This list is guaranteed to be
// complete and valid. The returned transactions
// are indexed by their position in the block.
func (p *Provider) GetTxsAtBlock(ctx context.Context, header *types.Header) ([]*TransactionWithIndex, error) {
	return p.tx.getTxsAtBlock(ctx, header)
}

// GetLogsAtBlock retrieves the logs for the specified
// Ethereum account at the specified block.
func (p *Provider) GetLogsAtBlock(ctx context.Context, acc common.Address, blockNum *big.Int) ([]*types.Log, error) {
	return p.log.getLogsAtBlock(ctx, acc, blockNum)
}

// GetAccountAtBlock provides the verified account
// at the specified block, or nil if no such account
// exists.
func (p *Provider) GetAccountAtBlock(ctx context.Context, acc common.Address, head *types.Header) (*Account, error) {
	return p.acc.getAccountAtBlock(ctx, acc, head)
}

// GetStorageAtBlock provides the verified value stored at
// the specified storage slot for the specified Ethereum
// account at the specified block.
//
// Note that the specified account must exist at the
// specified block, otherwise an error will be returned.
func (p *Provider) GetStorageAtBlock(ctx context.Context, acc common.Address, slot common.Hash, head *types.Header) ([]byte, error) {
	return p.acc.getSlotAtBlock(ctx, acc, slot, head)
}

// GetCodeAtBlock provides the verified code of the
// specified Ethereum account at the specified block.
//
// Note that the specified account must exist at the
// specified block, otherwise an error will be returned.
func (p *Provider) GetCodeAtBlock(ctx context.Context, acc common.Address, head *types.Header) ([]byte, error) {
	return p.acc.getCodeAtBlock(ctx, acc, head)
}

// CreateAccessList creates an access list for the
// specified transaction based on the state at the
// specified block number.
func (p *Provider) CreateAccessList(ctx context.Context, tx *TransactionWithSender, blockNum *big.Int) (*types.AccessList, error) {
	return p.tx.createAccessList(ctx, tx, blockNum)
}

package ethclient

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// RpcProvider is a wrapper for the Ethereum RPC API.
// Unlike Client, RpcProvider may add additional
// verification to the provided data.
type RpcProvider struct {
	tx  *txProvider
	log *logProvider
	acc *accountProvider
}

// NewRpcProvider creates a new RpcProvider that
// uses the specified Ethereum RPC client.
func NewRpcProvider(rpc *Client) *RpcProvider {
	return &RpcProvider{
		tx:  newTxProvider(rpc),
		log: newLogProvider(rpc),
		acc: newAccountProvider(rpc),
	}
}

// GetTxsAtBlock retrieves all transactions at the
// specified block. This list is guaranteed to be
// complete and valid. The returned transactions
// are indexed by their position in the block.
func (p *RpcProvider) GetTxsAtBlock(ctx context.Context, header *types.Header) ([]*TransactionWithIndex, error) {
	return p.tx.getTxsAtBlock(ctx, header)
}

// GetLogsAtBlock retrieves the logs for the specified
// Ethereum account at the specified block.
func (p *RpcProvider) GetLogsAtBlock(ctx context.Context, acc common.Address, blockNum *big.Int) ([]*types.Log, error) {
	return p.log.getLogsAtBlock(ctx, acc, blockNum)
}

// GetAccountAtBlock provides the verified account
// at the specified block, or nil if no such account
// exists.
func (p *RpcProvider) GetAccountAtBlock(ctx context.Context, acc common.Address, head *types.Header) (*Account, error) {
	return p.acc.getAccountAtBlock(ctx, acc, head)
}

// GetStorageAtBlock provides the verified value stored at
// the specified storage slot for the specified Ethereum
// account at the specified block.
//
// Note that the specified account must exist at the
// specified block, otherwise an error will be returned.
func (p *RpcProvider) GetStorageAtBlock(ctx context.Context, acc common.Address, slot common.Hash, head *types.Header) ([]byte, error) {
	return p.acc.getSlotAtBlock(ctx, acc, slot, head)
}

// GetCodeAtBlock provides the verified code of the
// specified Ethereum account at the specified block.
//
// Note that the specified account must exist at the
// specified block, otherwise an error will be returned.
func (p *RpcProvider) GetCodeAtBlock(ctx context.Context, acc common.Address, head *types.Header) ([]byte, error) {
	return p.acc.getCodeAtBlock(ctx, acc, head)
}

// GetTransactionTrace retrieves the transaction trace
// with a pre-state tracer for the specified transaction
// hash.
//
// The prestate tracer returns the accounts necessary to
// execute the specified transaction.
func (p *RpcProvider) GetTransactionTrace(ctx context.Context, txHash common.Hash) (*TransactionTrace, error) {
	return p.tx.getTransactionTrace(ctx, txHash)
}

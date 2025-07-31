package ethclient

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// Provider is an interface for retrieving
// verified Ethereum blockchain data.
type Provider interface {
	// GetTxsAtBlock retrieves all transactions at the
	// specified block. This list is guaranteed to be
	// complete and valid. The returned transactions
	// are indexed by their position in the block.
	GetTxsAtBlock(ctx context.Context, header *types.Header) ([]*TransactionWithIndex, error)

	// GetLogsAtBlock retrieves the logs for the specified
	// Ethereum account at the specified block.
	GetLogsAtBlock(ctx context.Context, acc common.Address, blockNum *big.Int) ([]*types.Log, error)

	// GetAccountAtBlock provides the verified account
	// at the specified block, or nil if no such account
	// exists.
	GetAccountAtBlock(ctx context.Context, acc common.Address, head *types.Header) (*Account, error)

	// GetStorageAtBlock provides the verified value stored at
	// the specified storage slot for the specified Ethereum
	// account at the specified block.
	//
	// Note that the specified account must exist at the
	// specified block, otherwise an error will be returned.
	GetStorageAtBlock(ctx context.Context, acc common.Address, slot common.Hash, head *types.Header) ([]byte, error)

	// GetCodeAtBlock provides the verified code of the
	// specified Ethereum account at the specified block.
	//
	// Note that the specified account must exist at the
	// specified block, otherwise an error will be returned.
	GetCodeAtBlock(ctx context.Context, acc common.Address, head *types.Header) ([]byte, error)

	// GetTransactionTrace retrieves the transaction trace
	// with a pre-state tracer for the specified transaction
	// hash.
	//
	// The prestate tracer returns the accounts necessary to
	// execute the specified transaction.
	//
	// Note that the returned trace is not verified, and hence
	// may not be complete or valid.
	GetTransactionTrace(ctx context.Context, txHash common.Hash) (*TransactionTrace, error)

	// CreateAccessList creates an access list for the
	// specified transaction based on the state at the
	// specified block number.
	CreateAccessList(ctx context.Context, tx *TransactionWithSender, blockNum *big.Int) (*types.AccessList, error)
}

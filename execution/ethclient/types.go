package ethclient

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// Account represents an Ethereum account. It wraps
// the verified account data (nonce, balance, code
// hash, and storage root) with the account address.
type Account struct {
	Address     common.Address
	Nonce       uint64
	Balance     *big.Int
	CodeHash    common.Hash
	StorageRoot common.Hash
}

// TransactionWithIndex wraps a transaction
// with its index in the block.
type TransactionWithIndex struct {
	Tx    *types.Transaction
	Index int
}

// TransactionWithSender wraps a transaction
// with the sender's (from) address.
type TransactionWithSender struct {
	Tx   *types.Transaction
	From common.Address
}

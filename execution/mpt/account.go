package mpt

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// Account represents an Ethereum account.
type Account struct {
	Nonce       uint64      `json:"nonce"`
	Balance     *big.Int    `json:"balance"`
	StorageRoot common.Hash `json:"storageRoot"`
	CodeHash    common.Hash `json:"codeHash"`
}

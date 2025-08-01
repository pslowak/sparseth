package ethclient

import (
	"encoding/json"
	"fmt"
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

// TransactionTrace represents a transaction trace
// that contains all accounts touched during the
// transaction execution.
type TransactionTrace struct {
	Accounts []*AccountTrace
}

func (t *TransactionTrace) UnmarshalJSON(data []byte) error {
	var rawTrace map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawTrace); err != nil {
		return err
	}

	for acc, rawFields := range rawTrace {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(rawFields, &fields); err != nil {
			return fmt.Errorf("failed to unmarshal fields of account %s: %w", acc, err)
		}
		trace := &AccountTrace{
			Address: common.HexToAddress(acc),
		}

		if rawStorage, exists := fields["storage"]; exists {
			var storage StorageTrace
			err := json.Unmarshal(rawStorage, &storage)
			if err != nil {
				return fmt.Errorf("failed to unmarshal storage for account %s: %w", acc, err)
			}

			trace.Storage = &storage
		} else {
			// No storage slots are touched
			trace.Storage = &StorageTrace{
				Slots: make([]common.Hash, 0),
			}
		}

		t.Accounts = append(t.Accounts, trace)
	}

	return nil
}

// AccountTrace represents an Ethereum account
// that was touched during a transaction trace.
type AccountTrace struct {
	Address common.Address
	Storage *StorageTrace
}

// StorageTrace represents the touched storage
// slots of an account during a transaction trace,
// the slots may be empty.
type StorageTrace struct {
	Slots []common.Hash
}

func (t *StorageTrace) UnmarshalJSON(data []byte) error {
	var rawSlots map[string]string
	if err := json.Unmarshal(data, &rawSlots); err != nil {
		return err
	}
	for slot := range rawSlots {
		t.Slots = append(t.Slots, common.HexToHash(slot))
	}
	return nil
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

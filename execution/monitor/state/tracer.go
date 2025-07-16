package state

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"sparseth/log"
)

// Tracer keeps track of accounts and storage
// slots that have been written to.
type Tracer struct {
	// accountWrites keeps track of accounts
	// that have been written to
	accountWrites map[common.Address]bool
	// storageWrites keeps track of storage slots
	// that have been written to for each account
	storageWrites map[common.Address]map[common.Hash]bool
	// log is the logger for the tracer
	log log.Logger
}

// NewTracer creates a new Tracer instance
// with the specified logger.
func NewTracer(log log.Logger) *Tracer {
	return &Tracer{
		accountWrites: make(map[common.Address]bool),
		storageWrites: make(map[common.Address]map[common.Hash]bool),
		log:           log.With("component", "state-tracer"),
	}
}

// OnReadAccount checks if the specified account address
// has been written to.
func (t *Tracer) OnReadAccount(addr common.Address) error {
	if !t.accountWrites[addr] {
		t.log.Debug("uninitialized account read", "account", addr.Hex())
		return fmt.Errorf("uninitialized account %s", addr.Hex())
	}
	return nil
}

// OnWriteAccount marks the specified account address
// as having been written to.
func (t *Tracer) OnWriteAccount(addr common.Address) {
	t.accountWrites[addr] = true
}

// Accounts returns a slice of all account addresses
// that have been written to during tracing.
func (t *Tracer) Accounts() []common.Address {
	accounts := make([]common.Address, 0, len(t.accountWrites))
	for addr := range t.accountWrites {
		accounts = append(accounts, addr)
	}
	return accounts
}

// OnReadStorage checks if the storage slot for the
// specified account address has been written to.
func (t *Tracer) OnReadStorage(addr common.Address, key common.Hash) error {
	if slots, exists := t.storageWrites[addr]; !exists || !slots[key] {
		t.log.Debug("uninitialized storage read", "account", addr.Hex(), "slot", key.Hex())
		return fmt.Errorf("uninitialized storage %s at %s", key.Hex(), addr.Hex())
	}
	return nil
}

// OnWriteStorage marks a storage slot as written to
// for the specified account address.
func (t *Tracer) OnWriteStorage(addr common.Address, key common.Hash) {
	if _, exists := t.storageWrites[addr]; !exists {
		t.storageWrites[addr] = make(map[common.Hash]bool)
	}
	t.storageWrites[addr][key] = true
}

// StorageSlots returns a slice of all storage slots
// that have been written to for the specified account.
func (t *Tracer) StorageSlots(addr common.Address) []common.Hash {
	if slots, exists := t.storageWrites[addr]; exists {
		keys := make([]common.Hash, 0, len(slots))
		for key := range slots {
			keys = append(keys, key)
		}
		return keys
	}
	return make([]common.Hash, 0)
}

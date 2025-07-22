package state

import (
	"github.com/ethereum/go-ethereum/common"
	"sparseth/log"
)

// StorageRead represents an uninitialized
// storage read, containing the account
// address and the slots that were read with
// no prior write operation.
type StorageRead struct {
	Address common.Address
	Slots   []common.Hash
}

// tracer keeps track of accounts and storage
// slots that have been written to.
type tracer struct {
	// accWrites keeps track of accounts
	// that have been written to
	accWrites map[common.Address]bool
	// uninitializedAccReads keeps track of
	// accounts that have been read from but not
	// written to in a prior operation, indicating
	// an uninitialized read.
	uninitializedAccReads map[common.Address]bool
	// storageWrites keeps track of storage slots
	// that have been written to for each account
	storageWrites map[common.Address]map[common.Hash]bool
	// uninitializedStorageReads keeps track of
	// storage slots that have been read from but
	// not written to in a prior operation,
	// indicating an uninitialized read.
	uninitializedStorageReads map[common.Address]map[common.Hash]bool
	// log is the logger for the tracer
	log log.Logger
}

// newTracer creates a new tracer instance
// with the specified logger.
func newTracer(log log.Logger) *tracer {
	return &tracer{
		accWrites:                 make(map[common.Address]bool),
		storageWrites:             make(map[common.Address]map[common.Hash]bool),
		uninitializedAccReads:     make(map[common.Address]bool),
		uninitializedStorageReads: make(map[common.Address]map[common.Hash]bool),
		log:                       log.With("component", "state-tracer"),
	}
}

// OnReadAccount registers a read on the specified
// account address.
func (t *tracer) OnReadAccount(addr common.Address) {
	if !t.accWrites[addr] {
		t.uninitializedAccReads[addr] = true
		t.log.Debug("uninitialized account read", "account", addr.Hex())
	}
}

// OnWriteAccount marks the specified account address
// as having been written to.
func (t *tracer) OnWriteAccount(addr common.Address) {
	t.accWrites[addr] = true
}

// Accounts returns a slice of all account addresses
// that have been written to during tracing.
func (t *tracer) Accounts() []common.Address {
	accounts := make([]common.Address, 0, len(t.accWrites))
	for addr := range t.accWrites {
		accounts = append(accounts, addr)
	}
	return accounts
}

// UninitializedAccountReads returns a slice of all account
// addresses that have been read from but not written to
// in a prior operation, indicating an uninitialized read.
func (t *tracer) UninitializedAccountReads() []common.Address {
	uninitialized := make([]common.Address, 0, len(t.uninitializedAccReads))
	for addr := range t.uninitializedAccReads {
		uninitialized = append(uninitialized, addr)
	}
	return uninitialized
}

// OnReadStorage registers a read on the specified
// storage slot for the specified account address.
func (t *tracer) OnReadStorage(addr common.Address, key common.Hash) {
	if slots, exists := t.storageWrites[addr]; !exists || !slots[key] {
		if _, exists = t.uninitializedStorageReads[addr]; !exists {
			t.uninitializedStorageReads[addr] = make(map[common.Hash]bool)
		}
		t.uninitializedStorageReads[addr][key] = true
		t.log.Debug("uninitialized storage read", "account", addr.Hex(), "slot", key.Hex())
	}
}

// OnWriteStorage marks a storage slot as written to
// for the specified account address.
func (t *tracer) OnWriteStorage(addr common.Address, key common.Hash) {
	if _, exists := t.storageWrites[addr]; !exists {
		t.storageWrites[addr] = make(map[common.Hash]bool)
	}
	t.storageWrites[addr][key] = true
}

// StorageSlots returns a slice of all storage slots
// that have been written to for the specified account.
func (t *tracer) StorageSlots(addr common.Address) []common.Hash {
	if slots, exists := t.storageWrites[addr]; exists {
		keys := make([]common.Hash, 0, len(slots))
		for key := range slots {
			keys = append(keys, key)
		}
		return keys
	}
	return make([]common.Hash, 0)
}

// UninitializedStorageReads returns a slice of all storage
// slots that have been read from but not written to in a
// prior operation, indicating an uninitialized read.
func (t *tracer) UninitializedStorageReads() []*StorageRead {
	reads := make([]*StorageRead, 0, len(t.uninitializedStorageReads))
	for addr, slots := range t.uninitializedStorageReads {
		keys := make([]common.Hash, 0, len(slots))
		for key := range slots {
			keys = append(keys, key)
		}
		reads = append(reads, &StorageRead{Address: addr, Slots: keys})
	}
	return reads
}

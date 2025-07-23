package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
)

// journalEntry records a change to the state that
// can be reverted later.
type journalEntry interface {
	// revert undoes the changes introduced
	// by this journal entry
	revert(db *state.StateDB)
}

// nonceChange records a change to an account's nonce.
type nonceChange struct {
	addr common.Address
	prev uint64
}

// revert undoes the nonce change.
func (n *nonceChange) revert(db *state.StateDB) {
	db.SetNonce(n.addr, n.prev, tracing.NonceChangeUnspecified)
}

// balanceChange records a change to an account's balance.
type balanceChange struct {
	addr common.Address
	prev *uint256.Int
}

// revert undoes the balance change.
func (b *balanceChange) revert(db *state.StateDB) {
	db.SetBalance(b.addr, b.prev, tracing.BalanceChangeUnspecified)
}

// codeChange records a change to an account's code.
type codeChange struct {
	addr common.Address
	prev []byte
}

// revert undoes the code change.
func (c *codeChange) revert(db *state.StateDB) {
	db.SetCode(c.addr, c.prev)
}

// storageChange records a change to an account's storage.
type storageChange struct {
	addr common.Address
	slot common.Hash
	prev common.Hash
}

// revert undoes the storage change.
func (s *storageChange) revert(db *state.StateDB) {
	db.SetState(s.addr, s.slot, s.prev)
}

// journal records a series of changes to the state.
type journal struct {
	entries []journalEntry
}

// emptyJournal creates a new empty journal.
func emptyJournal() *journal {
	return &journal{
		entries: make([]journalEntry, 0),
	}
}

// Reset clears the journal, removing all entries.
func (j *journal) Reset() {
	j.entries = j.entries[:0]
}

// Revert rewinds all changes made in the journal
// since the last reset.
func (j *journal) Revert(db *state.StateDB) {
	// Revert the journal entries in reverse order
	for i := len(j.entries) - 1; i >= 0; i-- {
		j.entries[i].revert(db)
	}
}

// NonceChange records a change to an account's nonce,
// capturing the previous nonce value.
func (j *journal) NonceChange(addr common.Address, prev uint64) {
	j.entries = append(j.entries, &nonceChange{
		addr: addr,
		prev: prev,
	})
}

// BalanceChange records a change to an account's balance,
// capturing the previous balance value.
func (j *journal) BalanceChange(addr common.Address, prev *uint256.Int) {
	j.entries = append(j.entries, &balanceChange{
		addr: addr,
		prev: prev.Clone(),
	})
}

// CodeChange records a change to an account's code,
// capturing the previous code bytes.
func (j *journal) CodeChange(addr common.Address, prev []byte) {
	j.entries = append(j.entries, &codeChange{
		addr: addr,
		prev: prev,
	})
}

// StorageChange records a change to an account's storage,
// capturing the previous value of the storage slot.
func (j *journal) StorageChange(addr common.Address, slot, prev common.Hash) {
	j.entries = append(j.entries, &storageChange{
		addr: addr,
		slot: slot,
		prev: prev,
	})
}

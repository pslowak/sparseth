package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
)

// RevertingStateDB wraps a state.StateDB with
// reverting capabilities. Unlike the standard
// state database, finalised changes can be
// reverted.
type RevertingStateDB struct {
	// inner is the underlying state.StateDB
	inner *state.StateDB
	// journal
	journal *journal
}

// NewRevertingStateDB creates a new reverting
// state database with the specified state root
// and backing database.
func NewRevertingStateDB(root common.Hash, db state.Database) (*RevertingStateDB, error) {
	i, err := state.New(root, db)
	if err != nil {
		return nil, err
	}
	j := emptyJournal()

	return &RevertingStateDB{
		inner:   i,
		journal: j,
	}, nil
}

// WithRoot creates a new state with
// the specified state root.
func (db *RevertingStateDB) WithRoot(root common.Hash) (*RevertingStateDB, error) {
	inner, err := state.New(root, db.inner.Database())
	if err != nil {
		return nil, err
	}

	return &RevertingStateDB{
		journal: db.journal,
		inner:   inner,
	}, nil
}

// Revert undoes all changes made to the
// state change since the last commit.
func (db *RevertingStateDB) Revert() {
	db.journal.Revert(db.inner)
}

//
// state.StateDB functions
//

func (db *RevertingStateDB) Commit(block uint64, deleteEmptyObjects bool, noStorageWiping bool) (common.Hash, error) {
	db.journal.Reset()
	return db.inner.Commit(block, deleteEmptyObjects, noStorageWiping)
}

func (db *RevertingStateDB) IntermediateRoot(deleteEmptyObjects bool) {
	db.inner.IntermediateRoot(deleteEmptyObjects)
}

func (db *RevertingStateDB) SetBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) {
	prev := db.inner.GetBalance(addr)
	db.journal.BalanceChange(addr, prev)

	db.inner.SetBalance(addr, amount, reason)
}

//
// vm.SateDB interface functions
//

func (db *RevertingStateDB) CreateAccount(addr common.Address) {
	db.inner.CreateAccount(addr)
}

func (db *RevertingStateDB) CreateContract(addr common.Address) {
	db.inner.CreateContract(addr)
}

func (db *RevertingStateDB) SubBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	return db.inner.SubBalance(addr, amount, reason)
}

func (db *RevertingStateDB) AddBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	return db.inner.AddBalance(addr, amount, reason)
}

func (db *RevertingStateDB) GetBalance(addr common.Address) *uint256.Int {
	return db.inner.GetBalance(addr)
}

func (db *RevertingStateDB) GetNonce(addr common.Address) uint64 {
	return db.inner.GetNonce(addr)
}

func (db *RevertingStateDB) SetNonce(addr common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	prev := db.inner.GetNonce(addr)
	db.journal.NonceChange(addr, prev)

	db.inner.SetNonce(addr, nonce, reason)
}

func (db *RevertingStateDB) GetCodeHash(addr common.Address) common.Hash {
	return db.inner.GetCodeHash(addr)
}

func (db *RevertingStateDB) GetCode(addr common.Address) []byte {
	return db.inner.GetCode(addr)
}

func (db *RevertingStateDB) SetCode(addr common.Address, code []byte) []byte {
	prev := db.inner.SetCode(addr, code)
	db.journal.CodeChange(addr, prev)

	return prev
}

func (db *RevertingStateDB) GetCodeSize(addr common.Address) int {
	return db.inner.GetCodeSize(addr)
}

func (db *RevertingStateDB) AddRefund(gas uint64) {
	db.inner.AddRefund(gas)
}

func (db *RevertingStateDB) SubRefund(gas uint64) {
	db.inner.SubRefund(gas)
}

func (db *RevertingStateDB) GetRefund() uint64 {
	return db.inner.GetRefund()
}

func (db *RevertingStateDB) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	return db.inner.GetCommittedState(addr, hash)
}

func (db *RevertingStateDB) GetState(addr common.Address, key common.Hash) common.Hash {
	return db.inner.GetState(addr, key)
}

func (db *RevertingStateDB) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	prev := db.inner.SetState(addr, key, value)
	db.journal.StorageChange(addr, key, prev)

	return prev
}

func (db *RevertingStateDB) GetStorageRoot(addr common.Address) common.Hash {
	return db.inner.GetStorageRoot(addr)
}

func (db *RevertingStateDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	return db.inner.GetTransientState(addr, key)
}

func (db *RevertingStateDB) SetTransientState(addr common.Address, key, value common.Hash) {
	db.inner.SetTransientState(addr, key, value)
}

func (db *RevertingStateDB) SelfDestruct(addr common.Address) uint256.Int {
	return db.inner.SelfDestruct(addr)
}

func (db *RevertingStateDB) HasSelfDestructed(addr common.Address) bool {
	return db.inner.HasSelfDestructed(addr)
}

func (db *RevertingStateDB) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	return db.inner.SelfDestruct6780(addr)
}

func (db *RevertingStateDB) Exist(addr common.Address) bool {
	return db.inner.Exist(addr)
}

func (db *RevertingStateDB) Empty(addr common.Address) bool {
	return db.inner.Empty(addr)
}

func (db *RevertingStateDB) AddressInAccessList(addr common.Address) bool {
	return db.inner.AddressInAccessList(addr)
}

func (db *RevertingStateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressPresent bool, slotPresent bool) {
	return db.inner.SlotInAccessList(addr, slot)
}

func (db *RevertingStateDB) AddAddressToAccessList(addr common.Address) {
	db.inner.AddAddressToAccessList(addr)
}

func (db *RevertingStateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	db.inner.AddSlotToAccessList(addr, slot)
}

func (db *RevertingStateDB) PointCache() *utils.PointCache {
	return db.inner.PointCache()
}

func (db *RevertingStateDB) Prepare(rules params.Rules, sender, coinbase common.Address, dst *common.Address, precompiles []common.Address, list types.AccessList) {
	db.inner.Prepare(rules, sender, coinbase, dst, precompiles, list)
}

func (db *RevertingStateDB) RevertToSnapshot(revid int) {
	db.inner.RevertToSnapshot(revid)
}

func (db *RevertingStateDB) Snapshot() int {
	return db.inner.Snapshot()
}

func (db *RevertingStateDB) AddLog(log *types.Log) {
	db.inner.AddLog(log)
}

func (db *RevertingStateDB) AddPreimage(hash common.Hash, bytes []byte) {
	db.inner.AddPreimage(hash, bytes)
}

func (db *RevertingStateDB) Witness() *stateless.Witness {
	return db.inner.Witness()
}

func (db *RevertingStateDB) AccessEvents() *state.AccessEvents {
	return db.inner.AccessEvents()
}

func (db *RevertingStateDB) Finalise(deleteEmptyObjects bool) {
	db.inner.Finalise(deleteEmptyObjects)
}

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
	"sparseth/log"
)

// TracingStateDB wraps a state.StateDB with
// tracing capabilities to detect uninitialized
// account and storage reads.
type TracingStateDB struct {
	// inner is the underlying state.StateDB
	inner *state.StateDB
	// tracer is used to track account and storage writes
	tracer *Tracer
	// log is the logger for the TracingStateDB
	log log.Logger
}

// NewWithEmptyTraces creates a new state with
// the specified state root and backing database.
//
// Note that the traces are empty.
func NewWithEmptyTraces(root common.Hash, db state.Database, log log.Logger) (*TracingStateDB, error) {
	tracer := NewTracer(log)

	inner, err := state.New(root, db)
	if err != nil {
		return nil, err
	}

	return &TracingStateDB{
		inner:  inner,
		tracer: tracer,
		log:    log.With("component", "tracing-state-db"),
	}, nil
}

// New creates a new state from the given state.
//
// Note that traces are preserved from the old
// state.
func New(root common.Hash, old *TracingStateDB) (*TracingStateDB, error) {
	inner, err := state.New(root, old.inner.Database())
	if err != nil {
		return nil, err
	}

	return &TracingStateDB{
		inner:  inner,
		tracer: old.tracer,
		log:    old.log,
	}, nil
}

// WrittenAccounts returns a slice of all addresses
// that have been written to during tracing.
func (db *TracingStateDB) WrittenAccounts() []common.Address {
	return db.tracer.Accounts()
}

// UninitializedAccountReads returns a slice of addresses
// that have been read from but not written to in a
// prior operation, indicating an uninitialized read.
//
// Note that reads are reset when NewWithEmptyTraces is called.
func (db *TracingStateDB) UninitializedAccountReads() []common.Address {
	return db.tracer.UninitializedAccountReads()
}

// UninitializedStorageReads returns a slice of all storage
// slots that have been read from but not written to in a
// prior operation, indicating an uninitialized read.
//
// Note that reads are reset when NewWithEmptyTraces is called.
func (db *TracingStateDB) UninitializedStorageReads() []*StorageRead {
	return db.tracer.UninitializedStorageReads()
}

// WrittenStorageSlots returns a slice of all storage slots
// that have been written to during tracing for the specified
// account address.
func (db *TracingStateDB) WrittenStorageSlots(addr common.Address) []common.Hash {
	return db.tracer.StorageSlots(addr)
}

//
// state.StateDB functions
//

func (db *TracingStateDB) Commit(block uint64, deleteEmptyObjects bool, noStorageWiping bool) (common.Hash, error) {
	return db.inner.Commit(block, deleteEmptyObjects, noStorageWiping)
}

func (db *TracingStateDB) GetLogs(thash common.Hash, bhash common.Hash, bNum uint64) []*types.Log {
	return db.inner.GetLogs(thash, bNum, bhash)
}

func (db *TracingStateDB) GetTrie() state.Trie {
	return db.inner.GetTrie()
}

func (db *TracingStateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return db.inner.IntermediateRoot(deleteEmptyObjects)
}

func (db *TracingStateDB) SetTxContext(thash common.Hash, ti int) {
	db.inner.SetTxContext(thash, ti)
}

func (db *TracingStateDB) SetBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) {
	db.inner.SetBalance(addr, amount, reason)
}

//
// vm.SateDB interface functions
//

func (db *TracingStateDB) CreateAccount(addr common.Address) {
	db.tracer.OnWriteAccount(addr)
	db.inner.CreateAccount(addr)
}

func (db *TracingStateDB) CreateContract(addr common.Address) {
	db.tracer.OnWriteAccount(addr)
	db.inner.CreateContract(addr)
}

func (db *TracingStateDB) SubBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	db.tracer.OnReadAccount(addr)
	return db.inner.SubBalance(addr, amount, reason)
}

func (db *TracingStateDB) AddBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	db.tracer.OnReadAccount(addr)
	return db.inner.AddBalance(addr, amount, reason)
}

func (db *TracingStateDB) GetBalance(addr common.Address) *uint256.Int {
	db.tracer.OnReadAccount(addr)
	return db.inner.GetBalance(addr)
}

func (db *TracingStateDB) GetNonce(addr common.Address) uint64 {
	db.tracer.OnReadAccount(addr)
	return db.inner.GetNonce(addr)
}

func (db *TracingStateDB) SetNonce(addr common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	db.inner.SetNonce(addr, nonce, reason)
}

func (db *TracingStateDB) GetCodeHash(addr common.Address) common.Hash {
	db.tracer.OnReadAccount(addr)
	return db.inner.GetCodeHash(addr)
}

func (db *TracingStateDB) GetCode(addr common.Address) []byte {
	db.tracer.OnReadAccount(addr)
	return db.inner.GetCode(addr)
}

func (db *TracingStateDB) SetCode(addr common.Address, code []byte) []byte {
	return db.inner.SetCode(addr, code)
}

func (db *TracingStateDB) GetCodeSize(addr common.Address) int {
	return db.inner.GetCodeSize(addr)
}

func (db *TracingStateDB) AddRefund(gas uint64) {
	db.inner.AddRefund(gas)
}

func (db *TracingStateDB) SubRefund(gas uint64) {
	db.inner.SubRefund(gas)
}

func (db *TracingStateDB) GetRefund() uint64 {
	return db.inner.GetRefund()
}

func (db *TracingStateDB) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	return db.inner.GetCommittedState(addr, hash)
}

func (db *TracingStateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	db.tracer.OnReadStorage(addr, hash)
	return db.inner.GetState(addr, hash)
}

func (db *TracingStateDB) SetState(addr common.Address, key, value common.Hash) common.Hash {
	db.tracer.OnWriteStorage(addr, key)
	return db.inner.SetState(addr, key, value)
}

func (db *TracingStateDB) GetStorageRoot(addr common.Address) common.Hash {
	db.tracer.OnReadAccount(addr)
	return db.inner.GetStorageRoot(addr)
}

func (db *TracingStateDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	return db.inner.GetTransientState(addr, key)
}

func (db *TracingStateDB) SetTransientState(addr common.Address, key, value common.Hash) {
	db.inner.SetTransientState(addr, key, value)
}

func (db *TracingStateDB) SelfDestruct(addr common.Address) uint256.Int {
	return db.inner.SelfDestruct(addr)
}

func (db *TracingStateDB) HasSelfDestructed(addr common.Address) bool {
	return db.inner.HasSelfDestructed(addr)
}

func (db *TracingStateDB) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	return db.inner.SelfDestruct6780(addr)
}

func (db *TracingStateDB) Exist(addr common.Address) bool {
	return db.inner.Exist(addr)
}

func (db *TracingStateDB) Empty(addr common.Address) bool {
	return db.inner.Empty(addr)
}

func (db *TracingStateDB) AddressInAccessList(addr common.Address) bool {
	return db.inner.AddressInAccessList(addr)
}

func (db *TracingStateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressPresent bool, slotPresent bool) {
	return db.inner.SlotInAccessList(addr, slot)
}

func (db *TracingStateDB) AddAddressToAccessList(addr common.Address) {
	db.inner.AddAddressToAccessList(addr)
}

func (db *TracingStateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	db.inner.AddSlotToAccessList(addr, slot)
}

func (db *TracingStateDB) PointCache() *utils.PointCache {
	return db.inner.PointCache()
}

func (db *TracingStateDB) Prepare(rules params.Rules, sender, coinbase common.Address, dst *common.Address, precompiles []common.Address, list types.AccessList) {
	db.inner.Prepare(rules, sender, coinbase, dst, precompiles, list)
}

func (db *TracingStateDB) RevertToSnapshot(revid int) {
	db.inner.RevertToSnapshot(revid)
}

func (db *TracingStateDB) Snapshot() int {
	return db.inner.Snapshot()
}

func (db *TracingStateDB) AddLog(log *types.Log) {
	db.inner.AddLog(log)
}

func (db *TracingStateDB) AddPreimage(hash common.Hash, preimage []byte) {
	db.inner.AddPreimage(hash, preimage)
}

func (db *TracingStateDB) Witness() *stateless.Witness {
	return db.inner.Witness()
}

func (db *TracingStateDB) AccessEvents() *state.AccessEvents {
	return db.inner.AccessEvents()
}

func (db *TracingStateDB) Finalise(deleteEmptyObjects bool) {
	db.inner.Finalise(deleteEmptyObjects)
}

package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
	"log/slog"
	"math/big"
	"sparseth/internal/log"
	"sparseth/storage/mem"
	"testing"
)

func TestTracingStateDB_NewWithEmptyTraces(t *testing.T) {
	t.Run("should create db without error", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		_, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("should create db with empty traces", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(world.UninitializedAccountReads()) != 0 {
			t.Errorf("expected no uninitialized account reads")
		}
		if len(world.UninitializedStorageReads()) != 0 {
			t.Errorf("expected no uninitialized storage reads")
		}
	})
}

func TestTracingStateDB_New(t *testing.T) {
	t.Run("should not reset tracer on new state", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		old, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		old.GetNonce(addr)

		root, err := old.Commit(0, false, false)
		if err != nil {
			t.Fatalf("expected no error on commit, got: %v", err)
		}

		world, err := New(root, old)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		reads := world.UninitializedAccountReads()
		if len(reads) != 1 {
			t.Errorf("expected uninitialized 1 read, got %d", len(reads))
		}
		if reads[0] != addr {
			t.Errorf("expected uninitialized read for %s, got %s", addr.Hex(), reads[0].Hex())
		}
	})
}

func TestTracingStateDB_UninitializedAccountReads(t *testing.T) {
	t.Run("should register uninitialized account read on SubBalance", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		world.SubBalance(addr, uint256.MustFromBig(big.NewInt(1)), tracing.BalanceChangeUnspecified)

		reads := world.UninitializedAccountReads()
		if len(reads) != 1 {
			t.Errorf("expected uninitialized 1 read, got %d", len(reads))
		}
		if reads[0] != addr {
			t.Errorf("expected uninitialized read for %s, got %s", addr.Hex(), reads[0].Hex())
		}
	})

	t.Run("should register uninitialized account read on AddBalance", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		world.AddBalance(addr, uint256.MustFromBig(big.NewInt(1)), tracing.BalanceChangeUnspecified)

		reads := world.UninitializedAccountReads()
		if len(reads) != 1 {
			t.Errorf("expected uninitialized 1 read, got %d", len(reads))
		}
		if reads[0] != addr {
			t.Errorf("expected uninitialized read for %s, got %s", addr.Hex(), reads[0].Hex())
		}
	})

	t.Run("should register uninitialized account read on GetBalance", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		world.GetBalance(addr)

		reads := world.UninitializedAccountReads()
		if len(reads) != 1 {
			t.Errorf("expected uninitialized 1 read, got %d", len(reads))
		}
		if reads[0] != addr {
			t.Errorf("expected uninitialized read for %s, got %s", addr.Hex(), reads[0].Hex())
		}
	})

	t.Run("should register uninitialized account read on GetNonce", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		world.GetNonce(addr)

		reads := world.UninitializedAccountReads()
		if len(reads) != 1 {
			t.Errorf("expected uninitialized 1 read, got %d", len(reads))
		}
		if reads[0] != addr {
			t.Errorf("expected uninitialized read for %s, got %s", addr.Hex(), reads[0].Hex())
		}
	})

	t.Run("should register uninitialized account read on GetCode", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		world.GetCode(addr)

		reads := world.UninitializedAccountReads()
		if len(reads) != 1 {
			t.Errorf("expected uninitialized 1 read, got %d", len(reads))
		}
		if reads[0] != addr {
			t.Errorf("expected uninitialized read for %s, got %s", addr.Hex(), reads[0].Hex())
		}
	})

	t.Run("should register uninitialized account read on GetCodeHash", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		world.GetCodeHash(addr)

		reads := world.UninitializedAccountReads()
		if len(reads) != 1 {
			t.Errorf("expected uninitialized 1 read, got %d", len(reads))
		}
		if reads[0] != addr {
			t.Errorf("expected uninitialized read for %s, got %s", addr.Hex(), reads[0].Hex())
		}
	})

	t.Run("should register uninitialized account read on GetStorageRoot", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		world.GetStorageRoot(addr)

		reads := world.UninitializedAccountReads()
		if len(reads) != 1 {
			t.Errorf("expected uninitialized 1 read, got %d", len(reads))
		}
		if reads[0] != addr {
			t.Errorf("expected uninitialized read for %s, got %s", addr.Hex(), reads[0].Hex())
		}
	})
}

func TestTracingStateDB_UninitializedStorageReads(t *testing.T) {
	t.Run("should register uninitialized storage read on GetState", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		slot := common.BigToHash(big.NewInt(0))
		world.GetState(addr, slot)

		reads := world.UninitializedStorageReads()
		if len(reads) != 1 {
			t.Errorf("expected uninitialized 1 read, got %d", len(reads))
		}
		if reads[0].Address != addr {
			t.Errorf("expected uninitialized read for %s, got %s", addr.Hex(), reads[0].Address.Hex())
		}
		if len(reads[0].Slots) != 1 {
			t.Errorf("expected 1 slot in uninitialized read, got %d", len(reads[0].Slots))
		}
		if reads[0].Slots[0] != slot {
			t.Errorf("expected uninitialized read for slot %s, got %s", slot.Hex(), reads[0].Slots[0].Hex())
		}
	})
}

func TestTracingStateDB_WrittenStorageSlots(t *testing.T) {
	t.Run("should return empty slice for account with no writes", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

		slots := world.WrittenStorageSlots(addr)
		if len(slots) != 0 {
			t.Errorf("expected no written storage slots, got %d", len(slots))
		}
	})

	t.Run("should return written storage slots for account", func(t *testing.T) {
		logger := log.New(slog.DiscardHandler)

		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		slot := common.BigToHash(big.NewInt(0))

		world.SetState(addr, slot, common.BigToHash(big.NewInt(1)))

		slots := world.WrittenStorageSlots(addr)
		if len(slots) != 1 {
			t.Errorf("expected 1 written storage slot, got %d", len(slots))
		}
		if slots[0] != slot {
			t.Errorf("expected written storage slot %s, got %s", slot.Hex(), slots[0].Hex())
		}
	})
}

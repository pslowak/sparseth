package state

import (
	"github.com/ethereum/go-ethereum/common"
	"log/slog"
	"math/big"
	"sparseth/internal/log"
	"testing"
)

func TestTracer_Accounts(t *testing.T) {
	t.Run("should return empty slice when no accounts written", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		accounts := tracer.Accounts()
		if len(accounts) != 0 {
			t.Errorf("expected empty accounts slice, got: %v", accounts)
		}
	})

	t.Run("should return accounts that have been written to", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		first := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		tracer.OnWriteAccount(first)
		second := common.HexToAddress("0xcafecafecafecafecafecafecafecafecafecafe")
		tracer.OnWriteAccount(second)

		accounts := tracer.Accounts()
		if len(accounts) != 2 {
			t.Errorf("expected 2 accounts, got: %d", len(accounts))
		}

		// Check if both accounts are present
		if accounts[0] != first && accounts[1] != first {
			t.Errorf("expected account %s, got: %s", first.Hex(), accounts[0].Hex())
		}
		if accounts[0] != second && accounts[1] != second {
			t.Errorf("expected account %s, got: %s", second.Hex(), accounts[0].Hex())
		}
	})
}

func TestTracer_UninitializedAccounts(t *testing.T) {
	t.Run("should return empty slice if no reads", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		if len(tracer.UninitializedAccounts()) != 0 {
			t.Errorf("expected empty uninitialized accounts slice, got: %v", tracer.UninitializedAccounts())
		}
	})

	t.Run("should return empty slice if no uninitialized reads", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		tracer.OnWriteAccount(addr)

		if len(tracer.UninitializedAccounts()) != 0 {
			t.Errorf("expected empty uninitialized accounts slice, got: %v", tracer.UninitializedAccounts())
		}
	})

	t.Run("should contain uninitialized account", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		tracer.OnReadAccount(addr)

		uninitialized := tracer.UninitializedAccounts()
		if len(uninitialized) != 1 {
			t.Errorf("expected 1 uninitialized account, got: %d", len(uninitialized))
		}
		if uninitialized[0] != addr {
			t.Errorf("expected uninitialized account %s, got: %s", addr.Hex(), uninitialized[0].Hex())
		}
	})
}

func TestTracer_StorageSlots(t *testing.T) {
	t.Run("should return empty slice when no storage slots written", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		slots := tracer.StorageSlots(addr)
		if len(slots) != 0 {
			t.Errorf("expected empty storage slots slice, got: %v", slots)
		}
	})

	t.Run("should return storage slots that have been written to", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		acc := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		first := common.BigToHash(big.NewInt(0))
		tracer.OnWriteStorage(acc, first)
		second := common.BigToHash(big.NewInt(1))
		tracer.OnWriteStorage(acc, second)

		slots := tracer.StorageSlots(acc)
		if len(slots) != 2 {
			t.Errorf("expected 2 storage slots, got: %d", len(slots))
		}

		// Check if both slots are present
		if slots[0] != first && slots[1] != first {
			t.Errorf("expected slot %s, got: %s", first.Hex(), slots[0].Hex())
		}
		if slots[0] != second && slots[1] != second {
			t.Errorf("expected slot %s, got: %s", second.Hex(), slots[0].Hex())
		}
	})
}

func TestTracer_UninitializedStorageSlots(t *testing.T) {
	t.Run("should return empty slice if no reads", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		if len(tracer.UninitializedStorageSlots()) != 0 {
			t.Errorf("expected empty uninitialized storage slots slice, got: %v", tracer.UninitializedStorageSlots())
		}
	})

	t.Run("should return empty slice if no uninitialized reads", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		tracer.OnWriteStorage(addr, common.BigToHash(big.NewInt(0)))

		if len(tracer.UninitializedStorageSlots()) != 0 {
			t.Errorf("expected empty uninitialized storage slots slice, got: %v", tracer.UninitializedStorageSlots())
		}
	})

	t.Run("should contain uninitialized storage slot", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		key := common.BigToHash(big.NewInt(0))
		tracer.OnReadStorage(addr, key)

		uninitialized := tracer.UninitializedStorageSlots()
		if len(uninitialized) != 1 {
			t.Errorf("expected 1 uninitialized storage slot, got: %d", len(uninitialized))
		}
		if uninitialized[0].Address != addr {
			t.Errorf("expected uninitialized storage slot of address %s, got: %s", addr.Hex(), uninitialized[0].Address.Hex())
		}
		if len(uninitialized[0].Slots) != 1 {
			t.Errorf("expected 1 slot in uninitialized storage slot, got: %d", len(uninitialized[0].Slots))
		}
	})
}

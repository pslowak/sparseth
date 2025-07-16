package state

import (
	"github.com/ethereum/go-ethereum/common"
	"log/slog"
	"math/big"
	"sparseth/internal/log"
	"testing"
)

func TestTracer_OnReadAccount(t *testing.T) {
	t.Run("should return error on uninitialized account read", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		if err := tracer.OnReadAccount(addr); err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should return no error on initialized account read", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		tracer.OnWriteAccount(addr)

		if err := tracer.OnReadAccount(addr); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}

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

func TestTracer_OnReadStorage(t *testing.T) {
	t.Run("should return error on uninitialized storage read (no account)", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		if err := tracer.OnReadStorage(addr, common.BigToHash(big.NewInt(0))); err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should return error on uninitialized storage read (no slot)", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		tracer.OnWriteStorage(addr, common.BigToHash(big.NewInt(0)))

		if err := tracer.OnReadStorage(addr, common.BigToHash(big.NewInt(1))); err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should return no error on initialized storage read", func(t *testing.T) {
		tracer := NewTracer(log.New(slog.DiscardHandler))

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		key := common.BigToHash(big.NewInt(0))
		tracer.OnWriteStorage(addr, key)

		if err := tracer.OnReadStorage(addr, key); err != nil {
			t.Errorf("expected no error, got: %v", err)
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

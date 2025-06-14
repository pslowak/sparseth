package ethstore

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"sparseth/storage/mem"
	"testing"
)

func TestHeaderStore_Put(t *testing.T) {
	t.Run("should store header without error", func(t *testing.T) {
		db := mem.New()
		defer db.Close()

		store := NewHeaderStore(db)
		header := &types.Header{
			Number: big.NewInt(0),
			Extra:  []byte("I am a test header"),
		}

		if err := store.Put(header); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestHeaderStore_GetByHash(t *testing.T) {
	t.Run("should return error when header not found", func(t *testing.T) {
		db := mem.New()
		defer db.Close()

		store := NewHeaderStore(db)
		if _, err := store.GetByHash(common.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")); err == nil {
			t.Errorf("expected error when header not found, got nil")
		}
	})

	t.Run("should return previously stored header", func(t *testing.T) {
		db := mem.New()
		defer db.Close()

		store := NewHeaderStore(db)
		header := &types.Header{
			Number: big.NewInt(1),
			Extra:  []byte("I am a test header"),
		}

		if err := store.Put(header); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		res, err := store.GetByHash(header.Hash())
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if res.Hash() != header.Hash() {
			t.Errorf("expected hash %s, got %s", header.Hash(), res.Hash())
		}
		if res.Number.Cmp(big.NewInt(1)) != 0 {
			t.Errorf("expected number 1, got %d", res.Number)
		}
		if !bytes.Equal(res.Extra, header.Extra) {
			t.Errorf("expected extra to be %s, got %s", header.Extra, res.Extra)
		}
	})
}

func TestHeaderStore_GetByNumber(t *testing.T) {
	t.Run("should return error when header not found", func(t *testing.T) {
		db := mem.New()
		defer db.Close()

		store := NewHeaderStore(db)
		if _, err := store.GetByNumber(1); err == nil {
			t.Errorf("expected error when header not found, got nil")
		}
	})

	t.Run("should return previously stored header", func(t *testing.T) {
		db := mem.New()
		defer db.Close()

		store := NewHeaderStore(db)
		header := &types.Header{
			Number: big.NewInt(1),
			Extra:  []byte("I am a test header"),
		}

		if err := store.Put(header); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		res, err := store.GetByNumber(1)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if res.Hash() != header.Hash() {
			t.Errorf("expected hash %s, got %s", header.Hash(), res.Hash())
		}
		if res.Number.Cmp(big.NewInt(1)) != 0 {
			t.Errorf("expected number 1, got %d", res.Number)
		}
		if !bytes.Equal(res.Extra, header.Extra) {
			t.Errorf("expected extra to be %s, got %s", header.Extra, res.Extra)
		}
	})
}

package ethstore

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"sparseth/storage/mem"
	"testing"
)

func TestEventStore_GetLog(t *testing.T) {
	t.Run("should return error when log not found", func(t *testing.T) {
		db := mem.New()
		defer db.Close()

		store := NewEventStore(db)
		if _, err := store.GetLog(common.BytesToHash([]byte("tx-1")), 1); err == nil {
			t.Errorf("should return error when log not found")
		}
	})

	t.Run("should return previously stored log", func(t *testing.T) {
		db := mem.New()
		defer db.Close()

		store := NewEventStore(db)
		logs := []*types.Log{
			{
				TxHash: common.BytesToHash([]byte("tx-1")),
				Index:  0,
				Data:   []byte("data-1"),
			},
			{
				TxHash: common.BytesToHash([]byte("tx-1")),
				Index:  1,
				Data:   []byte("data-2"),
			},
		}

		if err := store.PutAll(logs); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		log, err := store.GetLog(logs[0].TxHash, logs[0].Index)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !bytes.Equal(logs[0].Data, log.Data) {
			t.Errorf("expected %v, got %v", logs[0].Data, log.Data)
		}
	})
}

func TestEventStore_PutAll(t *testing.T) {
	t.Run("should store logs without error", func(t *testing.T) {
		db := mem.New()
		defer db.Close()

		store := NewEventStore(db)
		logs := []*types.Log{
			{
				TxHash: common.BytesToHash([]byte("tx-1")),
				Index:  0,
				Data:   []byte("data-1"),
			},
			{
				TxHash: common.BytesToHash([]byte("tx-1")),
				Index:  1,
				Data:   []byte("data-2"),
			},
		}

		if err := store.PutAll(logs); err != nil {
			t.Error("expected no error, got", err)
		}
	})
}

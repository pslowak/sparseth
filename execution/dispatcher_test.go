package execution

import (
	"github.com/ethereum/go-ethereum/core/types"
	"log/slog"
	"math/big"
	"sparseth/internal/log"
	"testing"
	"time"
)

func TestDispatcher_Subscribe(t *testing.T) {
	t.Run("should return same channel for same id", func(t *testing.T) {
		d := NewDispatcher(log.New(slog.DiscardHandler))

		this := d.Subscribe("id")
		that := d.Subscribe("id")

		if this != that {
			t.Errorf("expected %v, got %v", that, this)
		}
	})

	t.Run("should return different channel for different id", func(t *testing.T) {
		d := NewDispatcher(log.New(slog.DiscardHandler))

		this := d.Subscribe("this")
		that := d.Subscribe("that")

		if this == that {
			t.Errorf("expected different channel")
		}
	})
}

func TestDispatcher_Unsubscribe(t *testing.T) {
	t.Run("should close channel for id", func(t *testing.T) {
		d := NewDispatcher(log.New(slog.DiscardHandler))

		sub := d.Subscribe("sub")
		d.Unsubscribe("sub")

		_, open := <-sub
		if open {
			t.Errorf("expected closed channel")
		}
	})
}

func TestDispatcher_Close(t *testing.T) {
	t.Run("should close all channels on close", func(t *testing.T) {
		d := NewDispatcher(log.New(slog.DiscardHandler))

		first := d.Subscribe("first")
		second := d.Subscribe("second")

		d.Close()

		_, open := <-first
		if open {
			t.Errorf("expected closed channel")
		}

		_, open = <-second
		if open {
			t.Errorf("expected closed channel")
		}
	})
}

func TestDispatcher_Broadcast(t *testing.T) {
	t.Run("should broadcast head to all subscribers", func(t *testing.T) {
		d := NewDispatcher(log.New(slog.DiscardHandler))

		sub := d.Subscribe("sub")
		head := &types.Header{
			Number: big.NewInt(1),
		}
		d.Broadcast(head)

		select {
		case rcv := <-sub:
			if rcv.Number.Cmp(head.Number) != 0 {
				t.Errorf("expected %v, got %v", head, rcv)
			}
		case <-time.After(time.Second):
			t.Errorf("timeout: did not receive head")
		}
	})
}

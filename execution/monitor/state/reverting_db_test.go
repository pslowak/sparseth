package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/triedb"
	"math/big"
	"sparseth/storage/mem"
	"testing"
)

func TestRevertingStateDB_Revert(t *testing.T) {
	t.Run("should revert finalised changes", func(t *testing.T) {
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)

		world, err := NewRevertingStateDB(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("error creating revering state database: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		slot := common.BigToHash(big.NewInt(1))
		val := common.BigToHash(big.NewInt(2))
		world.SetState(addr, slot, val)

		if world.GetStorageRoot(addr) != types.EmptyRootHash {
			t.Fatalf("expected storage root to be empty, got %s", world.GetStorageRoot(addr).Hex())
		}

		// Call a finalizing operation
		world.IntermediateRoot(false)

		if world.GetStorageRoot(addr) == types.EmptyRootHash {
			t.Fatalf("expected storage root to be non-empty, got %s", world.GetStorageRoot(addr).Hex())
		}

		world.Revert()
		root, err := world.Commit(1, false, false)
		if err != nil {
			t.Fatalf("error committing reverted state: %v", err)
		}

		reverted, err := world.WithRoot(root)
		if err != nil {
			t.Fatalf("error creating new state with reverted root: %v", err)
		}

		if reverted.GetState(addr, slot) != (common.Hash{}) {
			t.Errorf("expected reverted state to be empty, got %s", reverted.GetState(addr, slot).Hex())
		}
		if reverted.GetStorageRoot(addr) != types.EmptyRootHash {
			t.Errorf("expected reverted state to be empty, got %s", reverted.GetStorageRoot(addr).Hex())
		}
	})
}

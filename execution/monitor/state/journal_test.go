package state

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
	"math/big"
	"sparseth/storage/mem"
	"testing"
)

func TestJournal_NonceChange(t *testing.T) {
	t.Run("should revert nonce change", func(t *testing.T) {
		world, err := createEmptyWorld()
		if err != nil {
			t.Fatalf("failed to create empty world state: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		prev := world.GetNonce(addr)

		nonce := uint64(42)
		world.SetNonce(addr, nonce, tracing.NonceChangeUnspecified)

		root, err := world.Commit(1, false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}
		newWorld, err := state.New(root, world.Database())

		// Record nonce change in journal
		j := emptyJournal()
		j.NonceChange(addr, prev)

		if newWorld.GetNonce(addr) != nonce {
			t.Fatalf("expected nonce %d, got %d", nonce, newWorld.GetNonce(addr))
		}

		// Revert nonce change
		j.Revert(newWorld)

		if newWorld.GetNonce(addr) != prev {
			t.Errorf("expected nonce %d, got %d", prev, newWorld.GetNonce(addr))
		}
	})
}

func TestJournal_BalanceChange(t *testing.T) {
	t.Run("should revert balance change", func(t *testing.T) {
		world, err := createEmptyWorld()
		if err != nil {
			t.Fatalf("failed to create empty world state: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		prev := world.GetBalance(addr)

		balance := uint256.MustFromBig(new(big.Int).Mul(big.NewInt(1), big.NewInt(params.Ether)))
		world.SetBalance(addr, balance, tracing.BalanceChangeUnspecified)

		root, err := world.Commit(1, false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}
		newWorld, err := state.New(root, world.Database())

		// Record nonce change in journal
		j := emptyJournal()
		j.BalanceChange(addr, prev)

		if !newWorld.GetBalance(addr).Eq(balance) {
			t.Fatalf("expected balance %d, got %d", balance, newWorld.GetBalance(addr))
		}

		// Revert balance change
		j.Revert(newWorld)

		if !newWorld.GetBalance(addr).Eq(prev) {
			t.Errorf("expected balance %d, got %d", prev, newWorld.GetBalance(addr))
		}
	})
}

func TestJournal_CodeChange(t *testing.T) {
	t.Run("should revert code change", func(t *testing.T) {
		world, err := createEmptyWorld()
		if err != nil {
			t.Fatalf("failed to create empty world state: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		prev := world.GetCode(addr)

		code := []byte("0x1234567890abcdef")
		world.SetCode(addr, code)

		root, err := world.Commit(1, false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}
		newWorld, err := state.New(root, world.Database())

		// Record nonce change in journal
		j := emptyJournal()
		j.CodeChange(addr, prev)

		if !bytes.Equal(newWorld.GetCode(addr), code) {
			t.Fatalf("expected code %d, got %d", code, newWorld.GetCode(addr))
		}
		if newWorld.GetCodeHash(addr) != crypto.Keccak256Hash(code) {
			t.Fatalf("expected code hash %s, got %s", crypto.Keccak256Hash(code).Hex(), newWorld.GetCodeHash(addr).Hex())
		}

		// Revert code change
		j.Revert(newWorld)

		if !bytes.Equal(newWorld.GetCode(addr), prev) {
			t.Errorf("expected code %d, got %d", prev, newWorld.GetCode(addr))
		}
		if newWorld.GetCodeHash(addr) != crypto.Keccak256Hash(prev) {
			t.Errorf("expected code hash %s, got %s", crypto.Keccak256Hash(prev).Hex(), newWorld.GetCodeHash(addr).Hex())
		}
	})
}

func TestJournal_StorageChange(t *testing.T) {
	t.Run("should revert storage change", func(t *testing.T) {
		world, err := createEmptyWorld()
		if err != nil {
			t.Fatalf("failed to create empty world state: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		slot := common.BigToHash(big.NewInt(1))
		prev := world.GetState(addr, slot)

		val := common.BigToHash(big.NewInt(2))
		world.SetState(addr, slot, val)

		if world.GetStorageRoot(addr) != types.EmptyRootHash {
			t.Fatalf("expected storage root to be empty, got %s", world.GetStorageRoot(addr).Hex())
		}

		updatedRoot, err := world.Commit(1, false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}
		updatedWorld, err := state.New(updatedRoot, world.Database())

		// Record nonce change in journal
		j := emptyJournal()
		j.StorageChange(addr, slot, prev)

		if updatedWorld.GetState(addr, slot) != val {
			t.Fatalf("expected state %s, got %s", val.Hex(), updatedWorld.GetState(addr, slot).Hex())
		}
		if updatedWorld.GetStorageRoot(addr) == types.EmptyRootHash {
			t.Fatalf("expected storage root to be non-empty, got %s", updatedWorld.GetStorageRoot(addr).Hex())
		}

		// Revert storage change
		j.Revert(updatedWorld)

		revertedRoot, err := updatedWorld.Commit(1, false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}
		revertedWorld, err := state.New(revertedRoot, updatedWorld.Database())
		if err != nil {
			t.Fatalf("failed to create reverted world state: %v", err)
		}

		if revertedWorld.GetState(addr, slot) != prev {
			t.Errorf("expected state %s, got %s", prev.Hex(), revertedWorld.GetState(addr, slot).Hex())
		}
		if revertedWorld.GetStorageRoot(addr) != types.EmptyRootHash {
			t.Errorf("expected storage root to be %s (empty), got %s", types.EmptyRootHash.Hex(), updatedWorld.GetStorageRoot(addr).Hex())
		}
	})
}

func TestJournal_Reset(t *testing.T) {
	t.Run("should not revert if reset before", func(t *testing.T) {
		world, err := createEmptyWorld()
		if err != nil {
			t.Fatalf("failed to create empty world state: %v", err)
		}

		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		slot := common.BigToHash(big.NewInt(1))
		prev := world.GetState(addr, slot)

		val := common.BigToHash(big.NewInt(2))
		world.SetState(addr, slot, val)

		if world.GetStorageRoot(addr) != types.EmptyRootHash {
			t.Fatalf("expected storage root to be empty, got %s", world.GetStorageRoot(addr).Hex())
		}

		updatedRoot, err := world.Commit(1, false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}
		updatedWorld, err := state.New(updatedRoot, world.Database())

		// Record nonce change in journal
		j := emptyJournal()
		j.StorageChange(addr, slot, prev)

		if updatedWorld.GetState(addr, slot) != val {
			t.Fatalf("expected state %s, got %s", val.Hex(), updatedWorld.GetState(addr, slot).Hex())
		}
		if updatedWorld.GetStorageRoot(addr) == types.EmptyRootHash {
			t.Fatalf("expected storage root to be non-empty, got %s", updatedWorld.GetStorageRoot(addr).Hex())
		}

		// Reset journal before reverting
		j.Reset()
		j.Revert(updatedWorld)

		revertedRoot, err := updatedWorld.Commit(1, false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}
		revertedWorld, err := state.New(revertedRoot, updatedWorld.Database())
		if err != nil {
			t.Fatalf("failed to create reverted world state: %v", err)
		}

		if revertedWorld.GetState(addr, slot) != val {
			t.Errorf("expected state %s, got %s", val.Hex(), revertedWorld.GetState(addr, slot).Hex())
		}
	})
}

func createEmptyWorld() (*state.StateDB, error) {
	db := rawdb.NewDatabase(mem.New())
	trieDB := triedb.NewDatabase(db, nil)
	stateDB := state.NewDatabase(trieDB, nil)
	return state.New(types.EmptyRootHash, stateDB)
}

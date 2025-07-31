package state

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
	"log/slog"
	"math/big"
	"sparseth/config"
	"sparseth/ethstore"
	"sparseth/execution/ethclient"
	"sparseth/internal/log"
	"sparseth/storage/mem"
	"testing"
)

type verifierTestProvider struct {
	// account to be retuned by GetAccountAtBlock
	acc *ethclient.Account
	// storage slot to be returned by GetStorageAtBlock
	storage []byte
	// error to be returned by provider methods
	err error
}

func (t verifierTestProvider) GetTxsAtBlock(context.Context, *types.Header) ([]*ethclient.TransactionWithIndex, error) {
	return nil, nil
}

func (t verifierTestProvider) GetLogsAtBlock(context.Context, common.Address, *big.Int) ([]*types.Log, error) {
	return nil, nil
}

func (t verifierTestProvider) GetAccountAtBlock(context.Context, common.Address, *types.Header) (*ethclient.Account, error) {
	return t.acc, t.err
}

func (t verifierTestProvider) GetStorageAtBlock(context.Context, common.Address, common.Hash, *types.Header) ([]byte, error) {
	return t.storage, t.err
}

func (t verifierTestProvider) GetCodeAtBlock(context.Context, common.Address, *types.Header) ([]byte, error) {
	return nil, nil
}

func (t verifierTestProvider) GetTransactionTrace(context.Context, common.Hash) (*ethclient.TransactionTrace, error) {
	return nil, nil
}

func (t verifierTestProvider) CreateAccessList(context.Context, *ethclient.TransactionWithSender, *big.Int) (*types.AccessList, error) {
	return nil, nil
}

func TestVerifier_VerifyUninitializedReads(t *testing.T) {
	t.Run("should return error when previous header cannot be retrieved", func(t *testing.T) {
		store := ethstore.NewHeaderStore(mem.New())
		v := NewVerifier(store, nil, log.New(slog.DiscardHandler))

		header := &types.Header{
			Number: big.NewInt(1),
		}
		if err := v.VerifyUninitializedReads(t.Context(), header, nil); err == nil {
			t.Errorf("exptected error when previous header cannot be retrieved, got nil")
		}
	})

	t.Run("should return no error when no uninitialized reads", func(t *testing.T) {
		prev := &types.Header{
			Number: big.NewInt(1),
		}

		store := ethstore.NewHeaderStore(mem.New())
		if err := store.Put(prev); err != nil {
			t.Fatalf("failed to store previous header: %v", err)
		}

		header := &types.Header{
			Number: big.NewInt(2),
		}

		logger := log.New(slog.DiscardHandler)
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("failed to create world state: %v", err)
		}

		v := NewVerifier(store, nil, logger)
		if err = v.VerifyUninitializedReads(t.Context(), header, world); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should return error when account could not be retrieved", func(t *testing.T) {
		prev := &types.Header{
			Number: big.NewInt(1),
		}

		store := ethstore.NewHeaderStore(mem.New())
		if err := store.Put(prev); err != nil {
			t.Fatalf("failed to store previous header: %v", err)
		}

		header := &types.Header{
			Number: big.NewInt(2),
		}

		logger := log.New(slog.DiscardHandler)
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("failed to create world state: %v", err)
		}

		// Create uninitialized read
		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		world.GetNonce(addr)

		testProvider := verifierTestProvider{
			acc: nil,
			err: fmt.Errorf("failed to retrieve account"),
		}

		v := NewVerifier(store, testProvider, logger)
		if err = v.VerifyUninitializedReads(t.Context(), header, world); err == nil {
			t.Errorf("expected error when account could not be retrieved, got nil")
		}
	})

	t.Run("should return error if account existed on-chain, but was flagged as uninitialized read", func(t *testing.T) {
		prev := &types.Header{
			Number: big.NewInt(1),
		}

		store := ethstore.NewHeaderStore(mem.New())
		if err := store.Put(prev); err != nil {
			t.Fatalf("failed to store previous header: %v", err)
		}

		header := &types.Header{
			Number: big.NewInt(2),
		}

		logger := log.New(slog.DiscardHandler)
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("failed to create world state: %v", err)
		}

		// Create uninitialized read
		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		world.GetNonce(addr)

		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address:     addr,
				Nonce:       1,
				Balance:     new(big.Int).Mul(big.NewInt(1), big.NewInt(params.GWei)),
				CodeHash:    types.EmptyCodeHash,
				StorageRoot: types.EmptyRootHash,
			},
		}

		v := NewVerifier(store, testProvider, logger)
		if err = v.VerifyUninitializedReads(t.Context(), header, world); err == nil {
			t.Errorf("expected error when account exists but was flagged as uninitialized read, got nil")
		}
	})

	t.Run("should return no error if account creation", func(t *testing.T) {
		prev := &types.Header{
			Number: big.NewInt(1),
		}

		store := ethstore.NewHeaderStore(mem.New())
		if err := store.Put(prev); err != nil {
			t.Fatalf("failed to store previous header: %v", err)
		}

		header := &types.Header{
			Number: big.NewInt(2),
		}

		logger := log.New(slog.DiscardHandler)
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("failed to create world state: %v", err)
		}

		// Create uninitialized read
		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		world.GetNonce(addr)

		v := NewVerifier(store, verifierTestProvider{}, logger)
		if err = v.VerifyUninitializedReads(t.Context(), header, world); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("should return no error if no or default value on-chain", func(t *testing.T) {
		prev := &types.Header{
			Number: big.NewInt(1),
		}

		store := ethstore.NewHeaderStore(mem.New())
		if err := store.Put(prev); err != nil {
			t.Fatalf("failed to store previous header: %v", err)
		}

		header := &types.Header{
			Number: big.NewInt(2),
		}

		logger := log.New(slog.DiscardHandler)
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		world, err := NewWithEmptyTraces(types.EmptyRootHash, stateDB, logger)
		if err != nil {
			t.Fatalf("failed to create world state: %v", err)
		}

		// Create uninitialized read
		addr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		slot := common.BigToHash(big.NewInt(1))
		world.GetState(addr, slot)

		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address:     addr,
				Nonce:       1,
				Balance:     new(big.Int).Mul(big.NewInt(1), big.NewInt(params.GWei)),
				CodeHash:    types.EmptyCodeHash,
				StorageRoot: types.EmptyRootHash,
			},
		}

		v := NewVerifier(store, testProvider, logger)
		if err = v.VerifyUninitializedReads(t.Context(), header, world); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}

func TestVerifier_VerifyCompleteness(t *testing.T) {
	t.Run("should return error when account cannot be retrieved", func(t *testing.T) {
		testProvider := verifierTestProvider{
			acc: nil,
			err: fmt.Errorf("failed to retrieve account"),
		}
		v := NewVerifier(nil, testProvider, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		world, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err == nil {
			t.Errorf("verifier should fail when account cannot be retrieved")
		}
	})

	t.Run("should succeed when account does not exist", func(t *testing.T) {
		v := NewVerifier(nil, verifierTestProvider{}, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		world, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err != nil {
			t.Errorf("verifier should succeed when no account")
		}
	})

	t.Run("should return error if account does not exist not in world state", func(t *testing.T) {
		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address: common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
			},
		}
		v := NewVerifier(nil, testProvider, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: testProvider.acc.Address,
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		world, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err == nil {
			t.Errorf("verifier should fail when account does not exist in world state")
		}
	})

	t.Run("should return error if nonce mismatch", func(t *testing.T) {
		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address: common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:   2,
			},
		}
		v := NewVerifier(nil, testProvider, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: testProvider.acc.Address,
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		old, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}
		old.CreateAccount(acc.Addr)
		old.SetNonce(acc.Addr, testProvider.acc.Nonce-1, tracing.NonceChangeUnspecified)
		root, err := old.Commit(head.Number.Uint64(), false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}

		world, err := state.New(root, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err == nil {
			t.Errorf("verifier should fail when nonce mismatch")
		}
	})

	t.Run("should return error if balance mismatch", func(t *testing.T) {
		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address: common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:   1,
				Balance: big.NewInt(1000),
			},
		}
		v := NewVerifier(nil, testProvider, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: testProvider.acc.Address,
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		old, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}
		old.CreateAccount(acc.Addr)
		old.SetNonce(acc.Addr, testProvider.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(big.NewInt(10)), tracing.BalanceChangeUnspecified)
		root, err := old.Commit(head.Number.Uint64(), false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}

		world, err := state.New(root, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err == nil {
			t.Errorf("verifier should fail when balance mismatch")
		}
	})

	t.Run("should return error if code hash mismatch", func(t *testing.T) {
		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address:  common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:    1,
				Balance:  big.NewInt(1000),
				CodeHash: common.HexToHash("0xdeadbeef"),
			},
		}
		v := NewVerifier(nil, testProvider, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: testProvider.acc.Address,
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		old, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}
		old.CreateAccount(acc.Addr)
		old.SetNonce(acc.Addr, testProvider.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(testProvider.acc.Balance), tracing.BalanceChangeUnspecified)
		old.SetCode(acc.Addr, []byte{0x01})

		root, err := old.Commit(head.Number.Uint64(), false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}

		world, err := state.New(root, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err == nil {
			t.Errorf("verifier should fail when code hash mismatch")
		}
	})

	t.Run("should return error if storage root mismatch", func(t *testing.T) {
		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address:     common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:       1,
				Balance:     big.NewInt(1000),
				CodeHash:    types.EmptyCodeHash,
				StorageRoot: common.HexToHash("0xdeadbeef"),
			},
		}
		v := NewVerifier(nil, testProvider, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: testProvider.acc.Address,
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		old, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}
		old.CreateAccount(acc.Addr)
		old.SetNonce(acc.Addr, testProvider.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(testProvider.acc.Balance), tracing.BalanceChangeUnspecified)
		old.SetCode(acc.Addr, []byte{})
		old.SetState(acc.Addr, common.BigToHash(big.NewInt(1)), common.BigToHash(big.NewInt(2)))

		root, err := old.Commit(head.Number.Uint64(), false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}

		world, err := state.New(root, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err == nil {
			t.Errorf("verifier should fail when storage root mismatch")
		}
	})

	t.Run("should succeed if valid EOA", func(t *testing.T) {
		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address:     common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:       1,
				Balance:     big.NewInt(1000),
				CodeHash:    types.EmptyCodeHash,
				StorageRoot: types.EmptyRootHash,
			},
		}
		v := NewVerifier(nil, testProvider, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr:           testProvider.acc.Address,
			ContractConfig: &config.ContractConfig{},
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		old, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}
		old.CreateAccount(acc.Addr)
		old.SetNonce(acc.Addr, testProvider.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(testProvider.acc.Balance), tracing.BalanceChangeUnspecified)
		old.SetCode(acc.Addr, []byte{})

		root, err := old.Commit(head.Number.Uint64(), false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}

		world, err := state.New(root, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err != nil {
			t.Errorf("verifier should succeed for valid EOA")
		}
	})

	t.Run("should return error if interaction counter mismatch", func(t *testing.T) {
		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address:     common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:       1,
				Balance:     big.NewInt(1000),
				CodeHash:    crypto.Keccak256Hash([]byte("0xdeadbeef")),
				StorageRoot: common.HexToHash("0xf38f9f63c760d088d7dd04f743619b6291f63beebd8bdf530628f90e9cfa52d7"),
			},
			storage: common.BigToHash(big.NewInt(2)).Bytes(),
		}
		v := NewVerifier(nil, testProvider, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: testProvider.acc.Address,
			ContractConfig: &config.ContractConfig{
				State: &config.SparseConfig{
					CountSlot: common.BigToHash(big.NewInt(1)),
				},
			},
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		old, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}
		old.CreateAccount(acc.Addr)
		old.SetNonce(acc.Addr, testProvider.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(testProvider.acc.Balance), tracing.BalanceChangeUnspecified)
		old.SetCode(acc.Addr, []byte("0xdeadbeef"))
		old.SetState(acc.Addr, acc.ContractConfig.State.CountSlot, common.BigToHash(big.NewInt(1)))

		root, err := old.Commit(head.Number.Uint64(), false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}

		world, err := state.New(root, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err == nil {
			t.Errorf("verifier should fail when interaction counter mismatch")
		}
	})

	t.Run("should succeed if valid contract account", func(t *testing.T) {
		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address:     common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:       1,
				Balance:     big.NewInt(1000),
				CodeHash:    crypto.Keccak256Hash([]byte("0xdeadbeef")),
				StorageRoot: common.HexToHash("0xf38f9f63c760d088d7dd04f743619b6291f63beebd8bdf530628f90e9cfa52d7"),
			},
			storage: common.BigToHash(big.NewInt(1)).Bytes(),
		}
		v := NewVerifier(nil, testProvider, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: testProvider.acc.Address,
			ContractConfig: &config.ContractConfig{
				State: &config.SparseConfig{
					CountSlot: common.BigToHash(big.NewInt(1)),
				},
			},
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		old, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}
		old.CreateAccount(acc.Addr)
		old.SetNonce(acc.Addr, testProvider.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(testProvider.acc.Balance), tracing.BalanceChangeUnspecified)
		old.SetCode(acc.Addr, []byte("0xdeadbeef"))
		old.SetState(acc.Addr, acc.ContractConfig.State.CountSlot, common.BigToHash(big.NewInt(1)))

		root, err := old.Commit(head.Number.Uint64(), false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}

		world, err := state.New(root, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err != nil {
			t.Errorf("verifier should succeed for valid contract account, got: %v", err)
		}
	})

	t.Run("should succeed if contract exists but count slot was not written yet (contract creation)", func(t *testing.T) {
		testProvider := verifierTestProvider{
			acc: &ethclient.Account{
				Address:     common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:       0,
				Balance:     big.NewInt(0),
				CodeHash:    crypto.Keccak256Hash([]byte("0xdeadbeef")),
				StorageRoot: types.EmptyRootHash,
			},
			storage: nil,
		}
		v := NewVerifier(nil, testProvider, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: testProvider.acc.Address,
			ContractConfig: &config.ContractConfig{
				State: &config.SparseConfig{
					CountSlot: common.BigToHash(big.NewInt(1)),
				},
			},
		}
		head := &types.Header{
			Number: big.NewInt(1),
		}
		db := rawdb.NewDatabase(mem.New())
		trieDB := triedb.NewDatabase(db, nil)
		stateDB := state.NewDatabase(trieDB, nil)
		old, err := state.New(types.EmptyRootHash, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}
		old.CreateAccount(acc.Addr)
		old.SetNonce(acc.Addr, testProvider.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(testProvider.acc.Balance), tracing.BalanceChangeUnspecified)
		old.SetCode(acc.Addr, []byte("0xdeadbeef"))

		root, err := old.Commit(head.Number.Uint64(), false, false)
		if err != nil {
			t.Fatalf("failed to commit state: %v", err)
		}

		world, err := state.New(root, stateDB)
		if err != nil {
			t.Fatalf("failed to create new state: %v", err)
		}

		err = v.VerifyCompleteness(t.Context(), acc, head, world)
		if err != nil {
			t.Errorf("verifier should succeed for non-existent contract account, got: %v", err)
		}
	})
}

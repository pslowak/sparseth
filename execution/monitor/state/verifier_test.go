package state

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
	"log/slog"
	"math/big"
	"sparseth/execution/ethclient"
	"sparseth/internal/config"
	"sparseth/internal/log"
	"sparseth/storage/mem"
	"testing"
)

type VerifierTestProvider struct {
}

func (t VerifierTestProvider) GetTxsAtBlock(ctx context.Context, header *types.Header) ([]*ethclient.TransactionWithIndex, error) {
	return nil, nil
}

func (t VerifierTestProvider) GetLogsAtBlock(ctx context.Context, acc common.Address, blockNum *big.Int) ([]*types.Log, error) {
	return nil, nil
}

func (t VerifierTestProvider) GetAccountAtBlock(ctx context.Context, acc common.Address, head *types.Header) (*ethclient.Account, error) {
	if acc == common.HexToAddress("0x0000000000000000000000000000000000000001") {
		return nil, nil
	}
	if acc == common.HexToAddress("0x0000000000000000000000000000000000002") {
		return &ethclient.Account{
			Address:     acc,
			Nonce:       0,
			Balance:     big.NewInt(0),
			CodeHash:    types.EmptyCodeHash,
			StorageRoot: types.EmptyRootHash,
		}, nil
	}
	if acc == common.HexToAddress("0x0000000000000000000000000000000000003") {
		return &ethclient.Account{
			Address:     acc,
			Nonce:       1,
			Balance:     big.NewInt(0),
			CodeHash:    common.HexToHash("0xf1885eda54b7a053318cd41e2093220dab15d65381b1157a3633a83bfd5c9239"),
			StorageRoot: common.HexToHash("0xf38f9f63c760d088d7dd04f743619b6291f63beebd8bdf530628f90e9cfa52d7"),
		}, nil
	}
	if acc == common.HexToAddress("0x0000000000000000000000000000000000004") {
		return &ethclient.Account{
			Address:     acc,
			Nonce:       1,
			Balance:     big.NewInt(0),
			CodeHash:    common.HexToHash("0xf1885eda54b7a053318cd41e2093220dab15d65381b1157a3633a83bfd5c9239"),
			StorageRoot: common.HexToHash("0xf38f9f63c760d088d7dd04f743619b6291f63beebd8bdf530628f90e9cfa52d7"),
		}, nil
	}
	return nil, fmt.Errorf("failed to retrieve account")
}

func (t VerifierTestProvider) GetStorageAtBlock(ctx context.Context, acc common.Address, slot common.Hash, head *types.Header) ([]byte, error) {
	if acc == common.HexToAddress("0x0000000000000000000000000000000000003") {
		return common.BigToHash(big.NewInt(2)).Bytes(), nil
	}
	if acc == common.HexToAddress("0x0000000000000000000000000000000000004") {
		return common.BigToHash(big.NewInt(1)).Bytes(), nil
	}
	return nil, nil
}

func (t VerifierTestProvider) GetCodeAtBlock(ctx context.Context, acc common.Address, head *types.Header) ([]byte, error) {
	return nil, nil
}

func (t VerifierTestProvider) CreateAccessList(ctx context.Context, tx *ethclient.TransactionWithSender, blockNum *big.Int) (*types.AccessList, error) {
	return nil, nil
}

func TestVerifier_VerifyCompleteness(t *testing.T) {
	t.Run("should return error when account cannot be retrieved", func(t *testing.T) {
		v := NewVerifier(VerifierTestProvider{}, log.New(slog.DiscardHandler))

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
		v := NewVerifier(VerifierTestProvider{}, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: common.HexToAddress("0x0000000000000000000000000000000000000001"),
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

	t.Run("should return error if account doe not exist not in world state", func(t *testing.T) {
		v := NewVerifier(VerifierTestProvider{}, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: common.HexToAddress("0x0000000000000000000000000000000000000002"),
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
		v := NewVerifier(VerifierTestProvider{}, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: common.HexToAddress("0x0000000000000000000000000000000000000002"),
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
		old.SetNonce(acc.Addr, 1, tracing.NonceChangeUnspecified)
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
		v := NewVerifier(VerifierTestProvider{}, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: common.HexToAddress("0x0000000000000000000000000000000000000002"),
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
		old.SetNonce(acc.Addr, 0, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(big.NewInt(1)), tracing.BalanceChangeUnspecified)
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
		v := NewVerifier(VerifierTestProvider{}, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: common.HexToAddress("0x0000000000000000000000000000000000000002"),
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
		old.SetNonce(acc.Addr, 0, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(big.NewInt(0)), tracing.BalanceChangeUnspecified)
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
		v := NewVerifier(VerifierTestProvider{}, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: common.HexToAddress("0x0000000000000000000000000000000000000002"),
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
		old.SetNonce(acc.Addr, 0, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(big.NewInt(0)), tracing.BalanceChangeUnspecified)
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
		v := NewVerifier(VerifierTestProvider{}, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr:           common.HexToAddress("0x0000000000000000000000000000000000000002"),
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
		old.SetNonce(acc.Addr, 0, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(big.NewInt(0)), tracing.BalanceChangeUnspecified)
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
		v := NewVerifier(VerifierTestProvider{}, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: common.HexToAddress("0x0000000000000000000000000000000000000003"),
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
		old.SetNonce(acc.Addr, 1, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(big.NewInt(0)), tracing.BalanceChangeUnspecified)
		old.SetCode(acc.Addr, []byte{0x01, 0x02, 0x03})
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
		v := NewVerifier(VerifierTestProvider{}, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: common.HexToAddress("0x0000000000000000000000000000000000000004"),
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
		old.SetNonce(acc.Addr, 1, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(big.NewInt(0)), tracing.BalanceChangeUnspecified)
		old.SetCode(acc.Addr, []byte{0x01, 0x02, 0x03})
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
			t.Errorf("verifier should succeed for valid contract account")
		}
	})
}

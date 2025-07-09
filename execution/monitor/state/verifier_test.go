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
	// account to be retuned by GetAccountAtBlock
	acc *ethclient.Account
	// storage slot to be returned by GetStorageAtBlock
	storage []byte
	// error to be returned by provider methods
	err error
}

func (t VerifierTestProvider) GetTxsAtBlock(ctx context.Context, header *types.Header) ([]*ethclient.TransactionWithIndex, error) {
	return nil, nil
}

func (t VerifierTestProvider) GetLogsAtBlock(ctx context.Context, acc common.Address, blockNum *big.Int) ([]*types.Log, error) {
	return nil, nil
}

func (t VerifierTestProvider) GetAccountAtBlock(ctx context.Context, acc common.Address, head *types.Header) (*ethclient.Account, error) {
	return t.acc, t.err
}

func (t VerifierTestProvider) GetStorageAtBlock(ctx context.Context, acc common.Address, slot common.Hash, head *types.Header) ([]byte, error) {
	return t.storage, t.err
}

func (t VerifierTestProvider) GetCodeAtBlock(ctx context.Context, acc common.Address, head *types.Header) ([]byte, error) {
	return nil, nil
}

func (t VerifierTestProvider) CreateAccessList(ctx context.Context, tx *ethclient.TransactionWithSender, blockNum *big.Int) (*types.AccessList, error) {
	return nil, nil
}

func TestVerifier_VerifyCompleteness(t *testing.T) {
	t.Run("should return error when account cannot be retrieved", func(t *testing.T) {
		p := VerifierTestProvider{
			acc: nil,
			err: fmt.Errorf("failed to retrieve account"),
		}
		v := NewVerifier(p, log.New(slog.DiscardHandler))

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
		p := VerifierTestProvider{
			acc: &ethclient.Account{
				Address: common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
			},
		}
		v := NewVerifier(p, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: p.acc.Address,
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
		p := VerifierTestProvider{
			acc: &ethclient.Account{
				Address: common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:   2,
			},
		}
		v := NewVerifier(p, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: p.acc.Address,
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
		old.SetNonce(acc.Addr, p.acc.Nonce-1, tracing.NonceChangeUnspecified)
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
		p := VerifierTestProvider{
			acc: &ethclient.Account{
				Address: common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:   1,
				Balance: big.NewInt(1000),
			},
		}
		v := NewVerifier(p, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: p.acc.Address,
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
		old.SetNonce(acc.Addr, p.acc.Nonce, tracing.NonceChangeUnspecified)
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
		p := VerifierTestProvider{
			acc: &ethclient.Account{
				Address:  common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:    1,
				Balance:  big.NewInt(1000),
				CodeHash: common.HexToHash("0xdeadbeef"),
			},
		}
		v := NewVerifier(p, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: p.acc.Address,
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
		old.SetNonce(acc.Addr, p.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(p.acc.Balance), tracing.BalanceChangeUnspecified)
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
		p := VerifierTestProvider{
			acc: &ethclient.Account{
				Address:     common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:       1,
				Balance:     big.NewInt(1000),
				CodeHash:    types.EmptyCodeHash,
				StorageRoot: common.HexToHash("0xdeadbeef"),
			},
		}
		v := NewVerifier(p, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: p.acc.Address,
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
		old.SetNonce(acc.Addr, p.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(p.acc.Balance), tracing.BalanceChangeUnspecified)
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
		p := VerifierTestProvider{
			acc: &ethclient.Account{
				Address:     common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:       1,
				Balance:     big.NewInt(1000),
				CodeHash:    types.EmptyCodeHash,
				StorageRoot: types.EmptyRootHash,
			},
		}
		v := NewVerifier(p, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr:           p.acc.Address,
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
		old.SetNonce(acc.Addr, p.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(p.acc.Balance), tracing.BalanceChangeUnspecified)
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
		p := VerifierTestProvider{
			acc: &ethclient.Account{
				Address:     common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:       1,
				Balance:     big.NewInt(1000),
				CodeHash:    crypto.Keccak256Hash([]byte("0xdeadbeef")),
				StorageRoot: common.HexToHash("0xf38f9f63c760d088d7dd04f743619b6291f63beebd8bdf530628f90e9cfa52d7"),
			},
			storage: common.BigToHash(big.NewInt(2)).Bytes(),
		}
		v := NewVerifier(p, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: p.acc.Address,
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
		old.SetNonce(acc.Addr, p.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(p.acc.Balance), tracing.BalanceChangeUnspecified)
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
		p := VerifierTestProvider{
			acc: &ethclient.Account{
				Address:     common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
				Nonce:       1,
				Balance:     big.NewInt(1000),
				CodeHash:    crypto.Keccak256Hash([]byte("0xdeadbeef")),
				StorageRoot: common.HexToHash("0xf38f9f63c760d088d7dd04f743619b6291f63beebd8bdf530628f90e9cfa52d7"),
			},
			storage: common.BigToHash(big.NewInt(1)).Bytes(),
		}
		v := NewVerifier(p, log.New(slog.DiscardHandler))

		acc := &config.AccountConfig{
			Addr: p.acc.Address,
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
		old.SetNonce(acc.Addr, p.acc.Nonce, tracing.NonceChangeUnspecified)
		old.SetBalance(acc.Addr, uint256.MustFromBig(p.acc.Balance), tracing.BalanceChangeUnspecified)
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
			t.Errorf("verifier should succeed for valid contract account")
		}
	})
}

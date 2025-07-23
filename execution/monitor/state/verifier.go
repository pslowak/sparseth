package state

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"sparseth/ethstore"
	"sparseth/execution/ethclient"
	"sparseth/internal/config"
	"sparseth/log"
)

// Verifier is responsible for verifying the
// completeness of the state of monitored accounts.
type Verifier struct {
	store    *ethstore.HeaderStore
	provider ethclient.Provider
	log      log.Logger
}

// NewVerifier creates a new Verifier instance.
func NewVerifier(store *ethstore.HeaderStore, provider ethclient.Provider, log log.Logger) *Verifier {
	return &Verifier{
		store:    store,
		provider: provider,
		log:      log.With("component", "state-verifier"),
	}
}

// VerifyUninitializedReads checks whether the uninitialized
// reads from the world state are valid.
func (v *Verifier) VerifyUninitializedReads(ctx context.Context, header *types.Header, world *TracingStateDB) error {
	prev, err := v.store.GetByNumber(header.Number.Uint64() - 1)
	if err != nil {
		return fmt.Errorf("failed to get previous header: %w", err)
	}

	for _, acc := range world.UninitializedAccountReads() {
		if err = v.verifyAccountRead(ctx, acc, prev); err != nil {
			return fmt.Errorf("uninitialized account read for %s: %w", acc.Hex(), err)
		}
	}

	for _, tuple := range world.UninitializedStorageReads() {
		if err = v.verifyStorageRead(ctx, tuple, prev); err != nil {
			return fmt.Errorf("uninitialized storage read for account %s: %w", tuple.Address.Hex(), err)
		}
	}

	return nil
}

// verifyAccountRead checks whether the specified
// account exist at the specified previous block,
// indicating an invalid uninitialized read.
func (v *Verifier) verifyAccountRead(ctx context.Context, acc common.Address, prev *types.Header) error {
	expected, err := v.provider.GetAccountAtBlock(ctx, acc, prev)
	if err != nil {
		return fmt.Errorf("failed to fetch account %s: %w", acc.Hex(), err)
	}
	if expected != nil {
		return fmt.Errorf("account exists at block %d", prev.Number)
	}
	return nil
}

// verifyStorageRead checks whether the specified
// storage slots for the specified account exist
// at the specified previous block, indicating an
// invalid uninitialized read.
func (v *Verifier) verifyStorageRead(ctx context.Context, tuple *StorageRead, prev *types.Header) error {
	expected, err := v.provider.GetAccountAtBlock(ctx, tuple.Address, prev)
	if err != nil {
		return fmt.Errorf("failed to fetch account %s: %w", tuple.Address.Hex(), err)
	}
	if expected != nil {
		for _, slot := range tuple.Slots {
			val, err := v.provider.GetStorageAtBlock(ctx, tuple.Address, slot, prev)
			if err != nil {
				return fmt.Errorf("failed to fetch storage slot %s for account %s: %w", slot.Hex(), tuple.Address.Hex(), err)
			}
			if common.BytesToHash(val) != (common.Hash{}) {
				return fmt.Errorf("slot %s has non-default value at block %d", slot.Hex(), prev.Number)
			}
		}
	}

	return nil
}

// VerifyCompleteness checks whether the state of the
// specified account is complete.
//
// For EOAs, the on-chain state is compared to actual
// account state in the state database.
//
// For contract accounts, the on-chain interaction counter
// is compared to the actual interaction counter in the
// state database.
//
// This function does not modify the world state.
func (v *Verifier) VerifyCompleteness(ctx context.Context, acc *config.AccountConfig, header *types.Header, world vm.StateDB) error {
	v.log.Debug("verify state completeness", "account", acc.Addr.Hex(), "blockNum", header.Number.Uint64(), "blockHash", header.Hash().Hex())

	expected, err := v.provider.GetAccountAtBlock(ctx, acc.Addr, header)
	if err != nil {
		return fmt.Errorf("failed to fetch account")
	}
	if expected == nil {
		v.log.Info("account does not exist", "account", acc.Addr.Hex(), "num", header.Number.Uint64(), "hash", header.Hash().Hex())
		return nil
	}

	if err = v.verifyExternallyOwnedAccount(expected, header, world); err != nil {
		return err
	}

	// In addition to basic EOA validation,
	// we verify the interaction counter for
	// contract accounts
	if acc.ContractConfig.HasSparseConfig() {
		counter, err := v.provider.GetStorageAtBlock(ctx, acc.Addr, acc.ContractConfig.State.CountSlot, header)
		if err != nil {
			return fmt.Errorf("failed to fetch interaction counter: %w", err)
		}

		actual := world.GetState(acc.Addr, acc.ContractConfig.State.CountSlot)
		if common.BytesToHash(counter) != actual {
			v.logWithContext("interaction counter mismatch", expected, header)
			return fmt.Errorf("interaction counter mismatch: expected: %s, got: %s", common.Bytes2Hex(counter), actual.Hex())
		}
	}

	return nil
}

// verifyExternallyOwnedAccount verifies the state of an
// externally owned account (EOA) against the world state.
func (v *Verifier) verifyExternallyOwnedAccount(expected *ethclient.Account, header *types.Header, world vm.StateDB) error {
	if !world.Exist(expected.Address) {
		v.logWithContext("account exists on-chain but not in world state", expected, header)
		return fmt.Errorf("account does not exist in world state, but on-chain")
	}

	nonce := world.GetNonce(expected.Address)
	if expected.Nonce != nonce {
		v.logWithContext("nonce mismatch", expected, header)
		return fmt.Errorf("nonce mismatch: expected: %d, got; %d", expected.Nonce, nonce)
	}

	balance := world.GetBalance(expected.Address).ToBig()
	if expected.Balance.Cmp(balance) != 0 {
		v.logWithContext("balance mismatch", expected, header)
		return fmt.Errorf("balance mismatch: expected: %d, got: %d", expected.Balance, balance)
	}

	codeHash := world.GetCodeHash(expected.Address)
	if expected.CodeHash != codeHash {
		v.logWithContext("code hash mismatch", expected, header)
		return fmt.Errorf("code hash mismatch: expected: %s, got: %s", expected.CodeHash.Hex(), codeHash.Hex())
	}

	storageRoot := world.GetStorageRoot(expected.Address)
	if expected.StorageRoot != storageRoot {
		v.logWithContext("storage root mismatch", expected, header)
		return fmt.Errorf("storage root mismatch: expected: %s, got: %s", expected.StorageRoot.Hex(), storageRoot.Hex())
	}

	return nil
}

// logWithContext logs a message with the account
// address and block context at warn level.
func (v *Verifier) logWithContext(msg string, acc *ethclient.Account, header *types.Header) {
	v.log.Warn(msg, "addr", acc.Address.Hex(), "blockNum", header.Number.Uint64(), "blockHash", header.Hash().Hex())
}

package state

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"sparseth/execution/ethclient"
	"sparseth/internal/config"
	"sparseth/internal/log"
)

// Verifier is responsible for verifying the
// completeness of the state of monitored accounts.
type Verifier struct {
	provider ethclient.Provider
	log      log.Logger
}

// NewVerifier creates a new Verifier instance.
func NewVerifier(provider ethclient.Provider, log log.Logger) *Verifier {
	return &Verifier{
		provider: provider,
		log:      log.With("component", "state-verifier"),
	}
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
		if !bytes.Equal(counter, actual.Bytes()) {
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
	v.log.Debug(msg, "addr", acc.Address.Hex(), "blockNum", header.Number.Uint64(), "blockHash", header.Hash().Hex())
}

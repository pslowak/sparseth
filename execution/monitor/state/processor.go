package state

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"sparseth/ethstore"
	"sparseth/execution/ethclient"
	"sparseth/internal/config"
	"sparseth/log"
	"sparseth/storage"
)

// TxProcessor downloads and re-executes
// transactions relevant to the monitored
// accounts.
//
// Unlike LogProcessor, only one instance
// of TxProcessor is used for all monitored
// accounts.
type TxProcessor struct {
	provider ethclient.Provider
	executor *TxExecutor
	preparer *Preparer
	verifier *Verifier
	accounts *config.AccountsConfig
	log      log.Logger
}

// NewTxProcessor creates a new TxProcessor.
func NewTxProcessor(accs *config.AccountsConfig, cc *params.ChainConfig, db storage.KeyValStore, rpc *ethclient.Client, log log.Logger) *TxProcessor {
	provider := ethclient.NewRpcProvider(rpc)
	store := ethstore.NewHeaderStore(db)
	preparer := NewPreparer(provider, store, cc)
	verifier := NewVerifier(provider, log)
	executor := NewTxExecutor(cc)

	return &TxProcessor{
		provider: provider,
		executor: executor,
		preparer: preparer,
		verifier: verifier,
		accounts: accs,
		log:      log.With("component", "transaction-processor"),
	}
}

// ProcessBlock processes the specified block header.
func (p *TxProcessor) ProcessBlock(ctx context.Context, head *types.Header) error {
	p.logWithContext("download txs for block", head)
	txs, err := p.provider.GetTxsAtBlock(ctx, head)
	if err != nil {
		return fmt.Errorf("failed to get txs at block %d: %w", head.Number.Uint64(), err)
	}

	p.logWithContext("prepare state for block", head)
	world, err := p.preparer.LoadState(ctx, head, txs)
	if err != nil {
		return fmt.Errorf("failed to load state for block %d: %w", head.Number.Uint64(), err)
	}

	p.logWithContext("state before execution "+string(world.Dump(nil)), head)
	_, err = p.executor.ExecuteTxs(head, txs, world)
	if err != nil {
		return fmt.Errorf("failed to execute txs for block %d: %w", head.Number.Uint64(), err)
	}

	root, err := world.Commit(head.Number.Uint64(), false, false)
	if err != nil {
		return fmt.Errorf("failed to commit state for block %d: %w", head.Number.Uint64(), err)
	}

	newWorld, err := state.New(root, world.Database())
	if err != nil {
		return fmt.Errorf("failed to create new state for block %d: %w", head.Number.Uint64(), err)
	}
	p.logWithContext("state after execution "+string(newWorld.Dump(nil)), head)

	for _, acc := range p.accounts.Accounts {
		if err = p.verifier.VerifyCompleteness(ctx, acc, head, newWorld); err != nil {
			p.log.Warn("failed to verify state for account", "account", acc.Addr.Hex(), "num", head.Number, "hash", head.Hash().Hex(), "error", err)
			return fmt.Errorf("failed to verify state for account %s at block %d: %w", acc.Addr.Hex(), head.Number.Uint64(), err)
		}
	}

	return nil
}

// logWithContext logs a message with
// block context at debug level.
func (p *TxProcessor) logWithContext(msg string, header *types.Header) {
	p.log.Debug(msg, "num", header.Number, "hash", header.Hash().Hex())
}

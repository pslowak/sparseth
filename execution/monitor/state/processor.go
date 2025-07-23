package state

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
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
	world    *RevertingStateDB
	accounts *config.AccountsConfig
	log      log.Logger
}

// NewTxProcessor creates a new TxProcessor.
func NewTxProcessor(accs *config.AccountsConfig, cc *params.ChainConfig, db storage.KeyValStore, rpc *ethclient.Client, log log.Logger) (*TxProcessor, error) {
	provider := ethclient.NewRpcProvider(rpc)

	store := ethstore.NewHeaderStore(db)
	preparer := NewPreparer(provider, store, accs, cc, log)

	executor := NewTxExecutor(cc)
	verifier := NewVerifier(store, provider, log)

	rawDB := rawdb.NewDatabase(db)
	trieDB := triedb.NewDatabase(rawDB, nil)
	stateDB := state.NewDatabase(trieDB, nil)

	// The world state includes the verified and complete
	// state of all monitored accounts.
	world, err := NewRevertingStateDB(types.EmptyRootHash, stateDB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize state: %w", err)
	}

	return &TxProcessor{
		provider: provider,
		executor: executor,
		preparer: preparer,
		verifier: verifier,
		world:    world,
		accounts: accs,
		log:      log.With("component", "transaction-processor"),
	}, nil
}

// ProcessBlock processes the specified block header.
func (p *TxProcessor) ProcessBlock(ctx context.Context, head *types.Header) error {
	p.logWithContext("download txs for block", head)
	txs, err := p.provider.GetTxsAtBlock(ctx, head)
	if err != nil {
		return fmt.Errorf("failed to get txs at block %d: %w", head.Number.Uint64(), err)
	}

	p.logWithContext("filter txs for block", head)
	relevantTxs, err := p.preparer.FilterTxs(ctx, head, txs)
	if err != nil {
		return fmt.Errorf("failed to filter txs for block %d: %w", head.Number.Uint64(), err)
	}
	p.logWithContext(fmt.Sprintf("got: %d txs, filtered: %d txs, remaining: %d txs", len(txs), len(txs)-len(relevantTxs), len(relevantTxs)), head)

	if len(relevantTxs) == 0 {
		p.logWithContext("no txs to process, skip re-execution", head)
		return nil
	}

	p.logWithContext("prepare state for block", head)
	transientWorld, err := p.preparer.LoadState(ctx, head, relevantTxs)
	if err != nil {
		return fmt.Errorf("failed to load partial transient state for block %d: %w", head.Number.Uint64(), err)
	}

	p.logWithContext("process transactions for block", head)
	_, err = p.executor.ExecuteTxs(head, relevantTxs, transientWorld)
	if err != nil {
		return fmt.Errorf("failed to execute txs for block %d: %w", head.Number.Uint64(), err)
	}

	transientRoot, err := transientWorld.Commit(head.Number.Uint64(), false, false)
	if err != nil {
		return fmt.Errorf("failed to commit state for block %d: %w", head.Number.Uint64(), err)
	}

	newTransientWorld, err := New(transientRoot, transientWorld)
	if err != nil {
		return fmt.Errorf("failed to create new transient state for block %d: %w", head.Number.Uint64(), err)
	}

	p.logWithContext("verify uninitialized reads for block", head)
	if err = p.verifier.VerifyUninitializedReads(ctx, head, newTransientWorld); err != nil {
		p.log.Warn("invalid uninitialized reads detected", "num", head.Number, "hash", head.Hash().Hex(), "error", err)
		return fmt.Errorf("invalid uninitialized reads for block %d: %w", head.Number.Uint64(), err)
	}

	p.logWithContext("merge transient state into persistent state", head)
	p.merge(newTransientWorld)

	p.world.IntermediateRoot(false)

	p.logWithContext("verify state for block", head)
	for _, acc := range p.accounts.Accounts {
		if err = p.verifier.VerifyCompleteness(ctx, acc, head, p.world); err != nil {
			p.log.Warn("failed to verify state for account, reverting state changes", "account", acc.Addr.Hex(), "num", head.Number, "hash", head.Hash().Hex(), "error", err)
			p.world.Revert()
			return fmt.Errorf("failed to verify state for account %s at block %d: %w", acc.Addr.Hex(), head.Number.Uint64(), err)
		}
	}

	p.logWithContext("verification succeeded, commit persistent state for block", head)
	root, err := p.world.Commit(head.Number.Uint64(), false, false)
	if err != nil {
		p.log.Warn("failed to commit persistent state for block", "num", head.Number, "hash", head.Hash().Hex(), "error", err)
		return fmt.Errorf("failed to commit persistent state for block %d: %w", head.Number.Uint64(), err)
	}

	p.world, err = p.world.WithRoot(root)
	if err != nil {
		p.log.Warn("failed to create new persistent state for block", "num", head.Number, "hash", head.Hash().Hex(), "error", err)
		return fmt.Errorf("failed to create new persistent state for block %d: %w", head.Number.Uint64(), err)
	}

	return nil
}

// logWithContext logs a message with
// block context at debug level.
func (p *TxProcessor) logWithContext(msg string, header *types.Header) {
	p.log.Debug(msg, "num", header.Number, "hash", header.Hash().Hex())
}

// merge merges the relevant changes from the transient
// world state ('from') into the persistent world state.
// A change is considered relevant if it affects a
// monitored account or its storage slots.
func (p *TxProcessor) merge(from *TracingStateDB) {
	// Merge accounts
	for _, acc := range from.WrittenAccounts() {
		if p.accounts.Contains(acc) {
			p.world.SetNonce(acc, from.GetNonce(acc), tracing.NonceChangeUnspecified)
			p.world.SetBalance(acc, from.GetBalance(acc), tracing.BalanceChangeUnspecified)
			p.world.SetCode(acc, from.GetCode(acc))
		}
	}

	// Merge storage slots
	for _, acc := range p.accounts.Accounts {
		for _, slot := range from.WrittenStorageSlots(acc.Addr) {
			val := from.GetState(acc.Addr, slot)
			p.world.SetState(acc.Addr, slot, val)
		}
	}
}

package state

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"sparseth/ethstore"
	"sparseth/execution/ethclient"
	"sparseth/internal/log"
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
	provider *ethclient.Provider
	executor *TxExecutor
	preparer *Preparer
	log      log.Logger
}

// NewTxProcessor creates a new TxProcessor.
func NewTxProcessor(cc *params.ChainConfig, db storage.KeyValStore, rpc *ethclient.Client, log log.Logger) *TxProcessor {
	provider := ethclient.NewProvider(rpc)
	store := ethstore.NewHeaderStore(db)
	preparer := NewPreparer(provider, store, cc)
	executor := NewTxExecutor(cc)

	return &TxProcessor{
		provider: provider,
		executor: executor,
		preparer: preparer,
		log:      log.With("component", "transaction-processor"),
	}
}

// ProcessBlock processes the specified block header.
func (p *TxProcessor) ProcessBlock(ctx context.Context, head *types.Header) error {
	p.log.Debug("download txs for block", "num", head.Number, "hash", head.Hash().Hex())
	txs, err := p.provider.GetTxsAtBlock(ctx, head)
	if err != nil {
		return fmt.Errorf("failed to get txs at block %d: %w", head.Number.Uint64(), err)
	}

	p.log.Debug("prepare state for block", "num", head.Number, "hash", head.Hash().Hex())
	world, err := p.preparer.LoadState(ctx, head, txs)
	if err != nil {
		return fmt.Errorf("failed to load state for block %d: %w", head.Number.Uint64(), err)
	}

	p.log.Debug("state before execution: "+string(world.Dump(nil)), "num", head.Number, "hash", head.Hash().Hex())

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

	p.log.Debug("state after execution: "+string(newWorld.Dump(nil)), "num", head.Number, "hash", head.Hash().Hex())
	return nil
}

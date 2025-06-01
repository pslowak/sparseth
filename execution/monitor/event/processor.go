package event

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"sparseth/ethstore"
	"sparseth/execution/ethclient"
	"sparseth/execution/monitor"
	"sparseth/internal/log"
)

// LogProcessor downloads, verifies and
// stores Ethereum event logs.
type LogProcessor struct {
	log      log.Logger
	acc      *monitor.AccountInfo
	verifier *Verifier
	db       *ethstore.EventStore
	provider *ethclient.Provider
}

// NewLogProcessor creates a new LogProcessor
// for the specified account.
func NewLogProcessor(acc *monitor.AccountInfo, rpc *ethclient.Client, store *ethstore.EventStore, log log.Logger) *LogProcessor {
	provider := ethclient.NewProvider(rpc)
	verifier := NewLogVerifier(acc.ABI, acc.InitialHead)

	return &LogProcessor{
		log:      log.With("component", acc.Addr.Hex()+"-log-processor"),
		acc:      acc,
		db:       store,
		provider: provider,
		verifier: verifier,
	}
}

// ProcessBlock processes the specified block header.
func (p *LogProcessor) ProcessBlock(ctx context.Context, head *types.Header) error {
	p.log.Debug("download logs for block", "num", head.Number, "hash", head.Hash().Hex())
	logs, err := p.provider.GetLogsAtBlock(ctx, p.acc.Addr, head.Number)
	if err != nil {
		return err
	}

	expected, err := p.provider.GetStorageAtBlock(ctx, p.acc.Addr, p.acc.Slot, head)
	if err != nil {
		return fmt.Errorf("failed to read header value: %w", err)
	}

	p.log.Debug("verify logs for block", "num", head.Number, "hash", head.Hash().Hex())
	if err = p.verifier.VerifyLogs(logs, common.BytesToHash(expected)); err != nil {
		return fmt.Errorf("failed to process logs: %w", err)
	}

	p.log.Debug("store logs for block", "num", head.Number, "hash", head.Hash().Hex())
	if err = p.db.PutAll(logs); err != nil {
		return fmt.Errorf("failed to store logs: %w", err)
	}

	p.log.Debug("block processed", "num", head.Number, "hash", head.Hash().Hex())
	return nil
}

package monitor

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"sparseth/ethstore"
	"sparseth/execution/ethclient"
	"sparseth/internal/log"
)

type Processor interface {
	// ProcessBlock handles a single block header.
	ProcessBlock(ctx context.Context, header *types.Header) error
}

// LogProcessor downloads, verifies and
// stores Ethereum event logs.
type LogProcessor struct {
	log      log.Logger
	acc      *AccountInfo
	verifier *Verifier
	db       *ethstore.EventStore
	rpc      *ethclient.Client
	reader   *ethclient.StorageReader
}

// NewLogProcessor creates a new LogProcessor
// for the specified account.
func NewLogProcessor(acc *AccountInfo, rpc *ethclient.Client, store *ethstore.EventStore, log log.Logger) *LogProcessor {
	reader := ethclient.NewStorageReader(rpc)
	verifier := NewLogVerifier(acc.ABI, acc.InitialHead)

	return &LogProcessor{
		log:      log.With("component", acc.Addr.Hex()+"-log-processor"),
		acc:      acc,
		rpc:      rpc,
		db:       store,
		reader:   reader,
		verifier: verifier,
	}
}

// ProcessBlock processes the specified block header.
func (p *LogProcessor) ProcessBlock(ctx context.Context, header *types.Header) error {
	p.log.Debug("download logs for block", "num", header.Number, "hash", header.Hash().Hex())
	logs, err := p.rpc.GetLogsAtBlock(ctx, p.acc.Addr, header.Number)
	if err != nil {
		return err
	}

	expected, err := p.reader.ReadSlot(ctx, p.acc.Addr, p.acc.Slot, header)
	if err != nil {
		return fmt.Errorf("failed to read header value: %w", err)
	}

	p.log.Debug("verify logs for block", "num", header.Number, "hash", header.Hash().Hex())
	if err = p.verifier.VerifyLogs(logs, common.BytesToHash(expected)); err != nil {
		return fmt.Errorf("failed to process logs: %w", err)
	}

	p.log.Debug("store logs for block", "num", header.Number, "hash", header.Hash().Hex())
	if err = p.db.PutAll(logs); err != nil {
		return fmt.Errorf("failed to store logs: %w", err)
	}

	p.log.Debug("block processed", "num", header.Number, "hash", header.Hash().Hex())
	return nil
}

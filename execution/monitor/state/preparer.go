package state

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
	"math/big"
	"sparseth/execution/ethclient"
	"sparseth/storage/mem"
)

// transactionWithContext wraps a transaction
// with its context, i.e., the access list and
// sender's address.
type transactionWithContext struct {
	tx         *types.Transaction
	accessList *types.AccessList
	sender     common.Address
}

// Preparer is responsible for preparing
// the partial world state just before
// the execution of a block.
type Preparer struct {
	provider *ethclient.Provider
	cc       *params.ChainConfig
}

// NewPreparer creates a new Preparer with the
// specified provider and chain configuration.
func NewPreparer(provider *ethclient.Provider, cc *params.ChainConfig) *Preparer {
	return &Preparer{
		provider: provider,
		cc:       cc,
	}
}

// LoadState reconstructs the partial state immediately before
// the specified block.
//
// In this context, 'partial state' refers to the state that is
// relevant for the execution of the provided transactions, i.e.,
// all accounts that are accessed by those transactions (including
// senders, recipients, and any account in their access lists).
// Unrelated accounts are omitted.
//
// Note that all transactions must belong to the specified block.
func (p *Preparer) LoadState(ctx context.Context, header *types.Header, txs []*ethclient.TransactionWithIndex) (*state.StateDB, error) {
	db := rawdb.NewDatabase(mem.New())
	trieDB := triedb.NewDatabase(db, nil)
	stateDB := state.NewDatabase(trieDB, nil)
	world, err := state.New(types.EmptyRootHash, stateDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create new state: %w", err)
	}

	prev, err := p.provider.GetHeaderByNumber(ctx, header.Number.Sub(header.Number, big.NewInt(1)))
	if err != nil {
		return nil, fmt.Errorf("failed to get previous header: %w", err)
	}

	txsWithAccessList, err := p.getTxsWithContext(ctx, header, txs)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions with access list: %w", err)
	}

	// Reconstruct the partial state
	// before the current block
	for _, t := range txsWithAccessList {
		if err = p.createStateForTx(ctx, prev, t, world); err != nil {
			return nil, fmt.Errorf("failed to create state for transaction at block %d: %w", prev.Number.Uint64(), err)
		}
	}

	root, err := world.Commit(prev.Number.Uint64(), false, false)
	if err != nil {
		return nil, fmt.Errorf("failed to commit state: %w", err)
	}

	return state.New(root, stateDB)
}

// getTxsWithContext retrieves the context for the
// specified transactions at the given block.
func (p *Preparer) getTxsWithContext(ctx context.Context, header *types.Header, txs []*ethclient.TransactionWithIndex) ([]*transactionWithContext, error) {
	result := make([]*transactionWithContext, len(txs))

	for i, tx := range txs {
		signer := types.MakeSigner(p.cc, header.Number, header.Time)
		from, err := signer.Sender(tx.Tx)
		if err != nil {
			return nil, fmt.Errorf("failed to get sender from tx at index %d: %w", i, err)
		}

		txWithSender := &ethclient.TransactionWithSender{
			Tx:   tx.Tx,
			From: from,
		}
		accessList, err := p.provider.CreateAccessList(ctx, txWithSender, header.Number)
		if err != nil {
			return nil, fmt.Errorf("failed to create access list for transaction %d: %w", i, err)
		}

		result[i] = &transactionWithContext{
			tx:         tx.Tx,
			accessList: accessList,
			sender:     from,
		}
	}

	return result, nil
}

// createStateForTx creates the relevant accounts
// for the specified transaction in the specified
// world state.
func (p *Preparer) createStateForTx(ctx context.Context, head *types.Header, tx *transactionWithContext, world *state.StateDB) error {
	// Create sender
	if err := p.createAccount(ctx, head, tx.sender, world); err != nil {
		return fmt.Errorf("failed to create sender account %s at block %d: %w", tx.sender.Hex(), head.Number.Uint64(), err)
	}

	// A nil receiver indicates a contract
	// creation transaction
	if tx.tx.To() != nil {
		if err := p.createAccount(ctx, head, *tx.tx.To(), world); err != nil {
			return fmt.Errorf("failed to create receiver account %s at block %d: %w", tx.tx.To().Hex(), head.Number.Uint64(), err)
		}
	}

	for _, tuple := range *tx.accessList {
		if err := p.createAccount(ctx, head, tuple.Address, world); err != nil {
			return fmt.Errorf("failed to create account %s at block %d: %w", tuple.Address.Hex(), head.Number.Uint64(), err)
		}

		for _, slot := range tuple.StorageKeys {
			// Initialize storage used by the tx
			if world.Exist(tuple.Address) {
				val, err := p.provider.GetStorageAtBlock(ctx, tuple.Address, slot, head)
				if err != nil {
					return fmt.Errorf("failed to get storage slot %s for account %s at block %d: %w", slot.Hex(), tuple.Address.Hex(), head.Number.Uint64(), err)
				}
				if val != nil {
					world.SetState(tuple.Address, slot, common.BytesToHash(val))
				}
			}
		}
	}

	return nil
}

// createAccount creates an account in the
// world state for the specified address.
// Note that storage is not initialized.
func (p *Preparer) createAccount(ctx context.Context, head *types.Header, addr common.Address, world *state.StateDB) error {
	if world.Exist(addr) {
		// Account already exists,
		// nothing to create
		return nil
	}

	acc, err := p.provider.GetAccountAtBlock(ctx, addr, head)
	if err != nil {
		return fmt.Errorf("failed to get account at block %d: %w", head.Number.Uint64(), err)
	}
	if acc == nil {
		// Account does not exist,
		// nothing to create
		return nil
	}

	world.CreateAccount(acc.Address)
	world.SetNonce(acc.Address, acc.Nonce, tracing.NonceChangeUnspecified)
	world.SetBalance(acc.Address, uint256.MustFromBig(acc.Balance), tracing.BalanceChangeUnspecified)

	if acc.CodeHash != types.EmptyCodeHash {
		code, err := p.provider.GetCodeAtBlock(ctx, acc.Address, head)
		if err != nil {
			return fmt.Errorf("failed to get code for account %s at block %d: %w", acc.Address.Hex(), head.Number.Uint64(), err)
		}
		world.SetCode(acc.Address, code)
	}

	return nil
}

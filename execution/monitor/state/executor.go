package state

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"sparseth/execution/ethclient"
)

// ExecutionResult contains the receipts
// generated during transaction execution.
type ExecutionResult struct {
	Receipts []*types.Receipt
}

// TxExecutor is responsible for executing
// transactions in the context of a block.
type TxExecutor struct {
	chain core.ChainContext
}

// NewTxExecutor creates a new TxExecutor
// using the supplied chain configuration.
// Note that TxExecutor is not safe for
// concurrent use.
func NewTxExecutor(chain *params.ChainConfig) *TxExecutor {
	return &TxExecutor{
		chain: &HeaderContext{
			Params: chain,
		},
	}
}

// ExecuteTxs executes the specified transactions
// using the supplied state. Not that it is assumed
// that all transactions belong to the supplied block.
func (e *TxExecutor) ExecuteTxs(header *types.Header, txs []*ethclient.TransactionWithIndex, state *state.StateDB) (*ExecutionResult, error) {
	usedGas := new(uint64)
	gasPool := new(core.GasPool).AddGas(header.GasLimit)

	signer := types.MakeSigner(e.chain.Config(), header.Number, header.Time)

	context := core.NewEVMBlockContext(header, e.chain, &header.Coinbase)
	evm := vm.NewEVM(context, state, e.chain.Config(), vm.Config{})

	receipts := make([]*types.Receipt, len(txs))
	for index, tx := range txs {
		msg, err := core.TransactionToMessage(tx.Tx, signer, header.BaseFee)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tx at index %d to message: %w", index, err)
		}
		state.SetTxContext(tx.Tx.Hash(), tx.Index)
		receipt, err := core.ApplyTransactionWithEVM(msg, gasPool, state, header.Number, header.Hash(), tx.Tx, usedGas, evm)
		if err != nil {
			return nil, fmt.Errorf("failed to apply transaction at index %d: %w", index, err)
		}
		receipts[index] = receipt
	}

	return &ExecutionResult{
		Receipts: receipts,
	}, nil
}

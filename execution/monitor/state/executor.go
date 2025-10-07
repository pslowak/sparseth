package state

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
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
func (e *TxExecutor) ExecuteTxs(header *types.Header, txs []*TransactionWithContext, world *TracingStateDB) (*ExecutionResult, error) {
	usedGas := new(uint64)
	gasPool := new(core.GasPool).AddGas(header.GasLimit)

	signer := types.MakeSigner(e.chain.Config(), header.Number, header.Time)

	context := core.NewEVMBlockContext(header, e.chain, &header.Coinbase)
	evm := vm.NewEVM(context, world, e.chain.Config(), vm.Config{})

	receipts := make([]*types.Receipt, len(txs))
	for index, tx := range txs {
		msg, err := core.TransactionToMessage(tx.Tx, signer, header.BaseFee)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tx at index %d to message: %w", index, err)
		}
		world.SetTxContext(tx.Tx.Hash(), tx.Index)

		onTxStart(evm, tx.Tx, msg)
		result, err := core.ApplyMessage(evm, msg, gasPool)
		if err != nil {
			onTxEnd(evm, nil, err)
			return nil, fmt.Errorf("failed to apply message at index %d: %w", index, err)
		}

		root := finalize(header.Number, evm, world)
		*usedGas += result.UsedGas

		if world.GetTrie().IsVerkle() {
			world.AccessEvents().Merge(evm.AccessEvents)
		}

		receipt := createReceipt(evm, result, world, header, tx, *usedGas, root)
		receipts[index] = receipt
		onTxEnd(evm, receipt, nil)
	}

	return &ExecutionResult{
		Receipts: receipts,
	}, nil
}

func onTxStart(evm *vm.EVM, tx *types.Transaction, msg *core.Message) {
	if hooks := evm.Config.Tracer; hooks != nil && hooks.OnTxStart != nil {
		hooks.OnTxStart(evm.GetVMContext(), tx, msg.From)
	}
}

func onTxEnd(evm *vm.EVM, receipt *types.Receipt, err error) {
	if hooks := evm.Config.Tracer; hooks != nil && hooks.OnTxEnd != nil {
		hooks.OnTxEnd(receipt, err)
	}
}

// finalize finalizes the state after executing
// a transaction in the block with the specified
// number.
func finalize(blockNum *big.Int, evm *vm.EVM, world *TracingStateDB) []byte {
	if evm.ChainConfig().IsByzantium(blockNum) {
		evm.StateDB.Finalise(true)
		return nil
	}

	return world.IntermediateRoot(evm.ChainConfig().IsEIP158(blockNum)).Bytes()
}

// createReceipt creates a receipt for the
// specified transaction execution result
// in the context of the specified block,
// EVM, and world state.
func createReceipt(evm *vm.EVM, result *core.ExecutionResult, world *TracingStateDB, header *types.Header, tx *TransactionWithContext, usedGas uint64, root []byte) *types.Receipt {
	status := types.ReceiptStatusSuccessful
	if result.Failed() {
		status = types.ReceiptStatusFailed
	}

	receipt := &types.Receipt{
		Status:            status,
		PostState:         root,
		Type:              tx.Tx.Type(),
		TxHash:            tx.Tx.Hash(),
		TransactionIndex:  uint(tx.Index),
		GasUsed:           result.UsedGas,
		BlockHash:         header.Hash(),
		BlockNumber:       header.Number,
		CumulativeGasUsed: usedGas,
	}

	if tx.Tx.Type() == types.BlobTxType {
		receipt.BlobGasUsed = uint64(len(tx.Tx.BlobHashes()) * params.BlobTxBlobGasPerBlob)
		receipt.BlobGasPrice = evm.Context.BlobBaseFee
	}

	if tx.Tx.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, tx.Tx.Nonce())
	}

	receipt.Logs = world.GetLogs(tx.Tx.Hash(), header.Hash(), header.Number.Uint64())
	receipt.Bloom = types.CreateBloom(receipt)
	return receipt
}

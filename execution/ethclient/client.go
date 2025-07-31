package ethclient

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"strings"
)

var (
	// prestateTracer is a tracer that returns
	// the accounts necessary to re-execute a
	// transaction.
	prestateTracer = map[string]string{
		"tracer": "prestateTracer",
	}
)

// Client is a wrapper for the
// Ethereum RPC API.
type Client struct {
	c *rpc.Client
}

// DialContext connects to an Ethereum
// RPC provider at the specified URL.
func DialContext(ctx context.Context, url string) (*Client, error) {
	c, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	return &Client{c: c}, nil
}

// NewClient creates a new Client instance
// using an existing RPC client connection.
func NewClient(c *rpc.Client) *Client {
	return &Client{c: c}
}

// Close shuts down the RPC client connection.
func (ec *Client) Close() error {
	ec.c.Close()
	return nil
}

// GetLogsAtBlock fetches the logs for the specified
// Ethereum account at the specified block.
func (ec *Client) GetLogsAtBlock(ctx context.Context, addr common.Address, blockNum *big.Int) ([]*types.Log, error) {
	type query struct {
		FromBlock string `json:"fromBlock"`
		ToBlock   string `json:"toBlock"`
		Address   string `json:"address"`
	}
	arg := &query{
		FromBlock: toBlockNumArg(blockNum),
		ToBlock:   toBlockNumArg(blockNum),
		Address:   addr.Hex(),
	}
	var result []*types.Log
	err := ec.c.CallContext(ctx, &result, "eth_getLogs", arg)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	return result, nil
}

// StorageProofEntry represents a proof
// for a key-value pair.
type StorageProofEntry struct {
	Key   common.Hash `json:"key"`
	Value []byte      `json:"value"`
	Proof [][]byte    `json:"proof"`
}

// Proof is the result of the GetProof operation.
type Proof struct {
	Address      common.Address       `json:"address"`
	Balance      *big.Int             `json:"balance"`
	Nonce        *big.Int             `json:"nonce"`
	CodeHash     common.Hash          `json:"codeHash"`
	StorageRoot  common.Hash          `json:"storageRoot"`
	AccountProof [][]byte             `json:"accountProof"`
	StorageProof []*StorageProofEntry `json:"storageProof"`
}

func (p *Proof) UnmarshalJSON(msg []byte) error {
	var raw struct {
		Address      common.Address       `json:"address"`
		Balance      *hexutil.Big         `json:"balance"`
		Nonce        *hexutil.Big         `json:"nonce"`
		CodeHash     common.Hash          `json:"codeHash"`
		StorageRoot  common.Hash          `json:"storageRoot"`
		AccountProof []string             `json:"accountProof"`
		StorageProof []*StorageProofEntry `json:"storageProof"`
	}

	if err := json.Unmarshal(msg, &raw); err != nil {
		return err
	}

	accProof, err := toProofNodes(raw.AccountProof)
	if err != nil {
		return err
	}

	p.Address = raw.Address
	p.Balance = raw.Balance.ToInt()
	p.Nonce = raw.Nonce.ToInt()
	p.CodeHash = raw.CodeHash
	p.StorageRoot = raw.StorageRoot
	p.AccountProof = accProof
	p.StorageProof = raw.StorageProof

	return nil
}

func (sp *StorageProofEntry) UnmarshalJSON(msg []byte) error {
	var raw struct {
		Key   common.Hash `json:"key"`
		Value string      `json:"value"`
		Proof []string    `json:"proof"`
	}

	if err := json.Unmarshal(msg, &raw); err != nil {
		return err
	}

	proof, err := toProofNodes(raw.Proof)
	if err != nil {
		return err
	}
	value, err := toByteSlice(raw.Value)
	if err != nil {
		return err
	}

	sp.Key = raw.Key
	sp.Proof = proof
	sp.Value = value

	return nil
}

// toProofNodes converts a slice of hex-encoded
// Merkle proof nodes into a slice of slices
// suitable for verification.
func toProofNodes(nodes []string) ([][]byte, error) {
	proofNodes := make([][]byte, len(nodes))
	for idx, node := range nodes {
		bytez, err := hex.DecodeString(strings.TrimPrefix(node, "0x"))
		if err != nil {
			return nil, fmt.Errorf("failed to decode node at index %d: %w", idx, err)
		}
		proofNodes[idx] = bytez
	}
	return proofNodes, nil
}

// toByteSlice converts a hex-encoded
// string into a byte slice.
func toByteSlice(val string) ([]byte, error) {
	if strings.HasPrefix(val, "0x") {
		val = strings.TrimPrefix(val, "0x")
	}
	if len(val)%2 != 0 {
		val = "0" + val
	}
	bytez, err := hex.DecodeString(val)
	if err != nil {
		return nil, fmt.Errorf("failed to decode byte array: %w", err)
	}
	return bytez, nil
}

// GetProof returns a Merkle proof for the specified
// storage slots of the specified account at the
// specified block. If the slots are nil or empty,
// the proof only contains the account proof.
func (ec *Client) GetProof(ctx context.Context, account common.Address, slots []common.Hash, blockHash common.Hash) (*Proof, error) {
	stringSlots := make([]string, len(slots))
	for i, s := range slots {
		stringSlots[i] = s.Hex()
	}
	var resp *Proof
	err := ec.c.CallContext(ctx, &resp, "eth_getProof", account.Hex(), stringSlots, blockHash.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get proof: %w", err)
	}
	return resp, nil
}

// GetCodeAtBlock retrieves the code for the specified
// Ethereum account at the specified block number.
func (ec *Client) GetCodeAtBlock(ctx context.Context, addr common.Address, blockNum *big.Int) ([]byte, error) {
	var code hexutil.Bytes
	err := ec.c.CallContext(ctx, &code, "eth_getCode", addr.Hex(), toBlockNumArg(blockNum))
	if err != nil {
		return nil, fmt.Errorf("failed to get code for address %s at block %s: %w", addr.Hex(), blockNum, err)
	}
	return code, nil
}

// GetTransactionsAtBlock retrieves all transactions
// from the block with the specified number.
func (ec *Client) GetTransactionsAtBlock(ctx context.Context, blockNum *big.Int) (types.Transactions, error) {
	type rpcBlock struct {
		Txs []*types.Transaction `json:"transactions"`
	}

	var block *rpcBlock
	err := ec.c.CallContext(ctx, &block, "eth_getBlockByNumber", toBlockNumArg(blockNum), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions at block %s: %w", blockNum, err)
	}
	if block == nil {
		return nil, fmt.Errorf("block %s not found", blockNum)
	}
	return block.Txs, err
}

// GetTransactionTrace retrieves the transaction trace
// with a pre-state tracer for the specified transaction
// hash.
//
// The prestate tracer returns the accounts necessary to
// execute the specified transaction.
func (ec *Client) GetTransactionTrace(ctx context.Context, txHash common.Hash) (*TransactionTrace, error) {
	var result *TransactionTrace
	err := ec.c.CallContext(ctx, &result, "debug_traceTransaction", txHash.Hex(), prestateTracer)
	if err != nil {
		return nil, fmt.Errorf("failed to trace transaction %s: %w", txHash.Hex(), err)
	}
	return result, nil
}

// CreateAccessList creates an access list for the
// specified transaction based on the state at the
// specified block number.
func (ec *Client) CreateAccessList(ctx context.Context, tx *types.Transaction, from common.Address, blockNum *big.Int) (*types.AccessList, error) {
	type req struct {
		From                 common.Address               `json:"from"`
		To                   *common.Address              `json:"to"`
		Value                *hexutil.Big                 `json:"value,omitempty"`
		GasPrice             *hexutil.Big                 `json:"gasPrice,omitempty"`
		MaxFeePerGas         *hexutil.Big                 `json:"maxFeePerGas,omitempty"`
		MaxPriorityFeePerGas *hexutil.Big                 `json:"maxPriorityFeePerGas,omitempty"`
		MaxFeePerBlobGas     *hexutil.Big                 `json:"maxFeePerBlobGas,omitempty"`
		BlobVersionedHashes  []common.Hash                `json:"blobVersionedHashes,omitempty"`
		AccessList           types.AccessList             `json:"accessList,omitempty"`
		AuthorizationList    []types.SetCodeAuthorization `json:"authorizationList,omitempty"`
		Input                hexutil.Bytes                `json:"input,omitempty"`
	}

	arg := &req{
		From: from,
		To:   tx.To(),
	}
	if val := tx.Value(); val != nil {
		arg.Value = (*hexutil.Big)(val)
	}
	if input := tx.Data(); len(input) > 0 {
		arg.Input = input
	}
	if gasPrice := tx.GasPrice(); gasPrice != nil {
		arg.GasPrice = (*hexutil.Big)(gasPrice)
	}
	if gasFeeCap := tx.GasFeeCap(); gasFeeCap != nil {
		arg.MaxFeePerGas = (*hexutil.Big)(gasFeeCap)
	}
	if gasTipCap := tx.GasTipCap(); gasTipCap != nil {
		arg.MaxPriorityFeePerGas = (*hexutil.Big)(gasTipCap)
	}
	if blobGasFeeCap := tx.BlobGasFeeCap(); blobGasFeeCap != nil {
		arg.MaxFeePerBlobGas = (*hexutil.Big)(blobGasFeeCap)
	}
	if blobHashes := tx.BlobHashes(); blobHashes != nil {
		arg.BlobVersionedHashes = blobHashes
	}
	if accessList := tx.AccessList(); accessList != nil {
		arg.AccessList = accessList
	}
	if authList := tx.SetCodeAuthorizations(); authList != nil {
		arg.AuthorizationList = authList
	}

	type rpcAccessList struct {
		AccessList *types.AccessList `json:"accessList"`
	}

	var accessList *rpcAccessList
	err := ec.c.CallContext(ctx, &accessList, "eth_createAccessList", arg, toBlockNumArg(blockNum))
	if err != nil {
		return nil, fmt.Errorf("failed to create access list: %w", err)
	}

	return accessList.AccessList, nil
}

// toBlockNumArg converts a *big.Int block number
// to a hex-encoded string suitable for RPC calls.
func toBlockNumArg(blockNum *big.Int) string {
	return fmt.Sprintf("0x%x", blockNum)
}

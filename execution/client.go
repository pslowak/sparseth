package execution

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"strings"
)

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

// Client is a wrapper for the
// Ethereum RPC API.
type Client struct {
	c *rpc.Client
}

// NewClient connects to an Ethereum RPC
// provider at the specified URL.
func NewClient(ctx context.Context, url string) (*Client, error) {
	c, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	return &Client{c: c}, nil
}

// Close shuts down the RPC client connection.
func (ec *Client) Close() error {
	ec.c.Close()
	return nil
}

// GetLogsAtBlock fetches the logs for the specified
// Ethereum account at the specified block.
func (ec *Client) GetLogsAtBlock(ctx context.Context, address common.Address, blockNumber *big.Int) ([]types.Log, error) {
	type query struct {
		FromBlock string `json:"fromBlock"`
		ToBlock   string `json:"toBlock"`
		Address   string `json:"address"`
	}
	arg := &query{
		FromBlock: fmt.Sprintf("0x%x", blockNumber),
		ToBlock:   fmt.Sprintf("0x%x", blockNumber),
		Address:   address.Hex(),
	}
	var result []types.Log
	err := ec.c.CallContext(ctx, &result, "eth_getLogs", arg)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	return result, nil
}

// GetProof returns a Merkle proof for the specified
// storage slots of the specified account at the
// specified block.
func (ec *Client) GetProof(ctx context.Context, account common.Address, slots []common.Hash, blockHash common.Hash) (*Proof, error) {
	type rpcStorageProofEntry struct {
		Key   string   `json:"key"`
		Value string   `json:"value"`
		Proof []string `json:"proof"`
	}
	type rpcProof struct {
		Address      string                  `json:"address"`
		Balance      string                  `json:"balance"`
		Code         string                  `json:"codeHash"`
		Nonce        string                  `json:"nonce"`
		StorageHash  string                  `json:"storageHash"`
		AccountProof []string                `json:"accountProof"`
		StorageProof []*rpcStorageProofEntry `json:"storageProof"`
	}

	slotHex := make([]string, len(slots))
	for i, s := range slots {
		slotHex[i] = s.Hex()
	}

	var resp rpcProof
	err := ec.c.CallContext(ctx, &resp, "eth_getProof", account.Hex(), slotHex, blockHash.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch proof: %w", err)
	}

	// Parse fields
	storageRoot := common.HexToHash(resp.StorageHash)
	address := common.HexToAddress(resp.Address)
	codeHash := common.HexToHash(resp.Code)

	balance := new(big.Int)
	balance.SetString(strings.TrimPrefix(resp.Balance, "0x"), 16)

	nonce := new(big.Int)
	nonce.SetString(strings.TrimPrefix(resp.Nonce, "0x"), 16)

	accountProof, err := toProofNodes(resp.AccountProof)
	if err != nil {
		return nil, err
	}

	storageProof := make([]*StorageProofEntry, len(resp.StorageProof))
	for i, entry := range resp.StorageProof {
		key := common.HexToHash(entry.Key)
		val, err := hex.DecodeString(strings.TrimPrefix(entry.Value, "0x"))
		if err != nil {
			return nil, fmt.Errorf("failed to decode value: %w", err)
		}
		proof, err := toProofNodes(entry.Proof)
		if err != nil {
			return nil, err
		}
		storageProof[i] = &StorageProofEntry{
			Key:   key,
			Value: val,
			Proof: proof,
		}
	}

	return &Proof{
		Address:      address,
		Balance:      balance,
		Nonce:        nonce,
		CodeHash:     codeHash,
		StorageRoot:  storageRoot,
		AccountProof: accountProof,
		StorageProof: storageProof,
	}, err
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

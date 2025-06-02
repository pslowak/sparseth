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
func (ec *Client) GetLogsAtBlock(ctx context.Context, address common.Address, blockNumber *big.Int) ([]*types.Log, error) {
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
	var result []*types.Log
	err := ec.c.CallContext(ctx, &result, "eth_getLogs", arg)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	return result, nil
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

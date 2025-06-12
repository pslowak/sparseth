package mpt

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"math/big"
	"sparseth/storage/mem"
)

// Account represents an Ethereum account.
type Account struct {
	Nonce       uint64      `json:"nonce"`
	Balance     *big.Int    `json:"balance"`
	StorageRoot common.Hash `json:"storageRoot"`
	CodeHash    common.Hash `json:"codeHash"`
}

// VerifyAccountProof verifies a Merkle proof for an Ethereum
// account against a given state root.
//
// If the account does not exist, but the proof is valid, nil
// is returned.
func VerifyAccountProof(stateRoot common.Hash, address common.Address, proofNodes [][]byte) (*Account, error) {
	key := crypto.Keccak256(address[:])
	data, err := verifyProof(stateRoot, key[:], proofNodes)
	if err != nil {
		return nil, err
	}
	if data == nil {
		// Non-existent account
		return nil, nil
	}

	var account Account
	if err := rlp.DecodeBytes(data, &account); err != nil {
		return nil, fmt.Errorf("failed to decode account: %w", err)
	}

	return &account, err
}

// VerifyStorageProof verifies a Merkle proof for a given slot key
// against a given storage root. If there is no value for the given
// slot key, nil is returned.
//
// Note that it is assumed that the slot key is a Keccak256 hash
// of the byte key.
func VerifyStorageProof(storageRoot common.Hash, slotKey common.Hash, proofNodes [][]byte) ([]byte, error) {
	if storageRoot == types.EmptyRootHash {
		// No storage for any key
		return nil, nil
	}

	data, err := verifyProof(storageRoot, slotKey[:], proofNodes)
	if err != nil {
		return nil, err
	}
	if data == nil {
		// No value for the given slot key
		return nil, nil
	}

	var val []byte
	if err := rlp.DecodeBytes(data, &val); err != nil {
		return nil, fmt.Errorf("failed to decode value: %w", err)
	}

	return val, nil
}

// verifyProof verifies a Merkle proof for a given key against
// a root hash.
//
// Note that the returned value is RLP encoded, or nil if no
// such value exists.
func verifyProof(rootHash common.Hash, key []byte, proofNodes [][]byte) ([]byte, error) {
	proof := mem.New()
	defer proof.Close()

	for _, node := range proofNodes {
		if err := proof.Put(crypto.Keccak256(node), node); err != nil {
			return nil, fmt.Errorf("failed to put proof node: %w", err)
		}
	}

	return trie.VerifyProof(rootHash, key, proof)
}

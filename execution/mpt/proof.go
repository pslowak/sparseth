package mpt

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"sparseth/execution/mpt/trienode"
)

// VerifyAccountProof verifies a Merkle proof for an Ethereum
// account against a given state root.
func VerifyAccountProof(stateRoot common.Hash, address common.Address, proofNodes [][]byte) (*Account, error) {
	key := crypto.Keccak256(address[:])
	data, err := verifyProof(stateRoot, key[:], proofNodes)
	if err != nil {
		return nil, err
	}

	var account Account
	if err := rlp.DecodeBytes(data, &account); err != nil {
		return nil, fmt.Errorf("failed to decode account: %w", err)
	}

	return &account, err
}

// VerifyStorageProof verifies a Merkle proof for a given slot key
// against a given storage root.
//
// Note that it is assumed that the slot key is a Keccak256 hash
// of the byte key.
func VerifyStorageProof(storageRoot common.Hash, slotKey common.Hash, proofNodes [][]byte) ([]byte, error) {
	data, err := verifyProof(storageRoot, slotKey[:], proofNodes)
	if err != nil {
		return nil, err
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
// Note that the returned value is RLP encoded.
func verifyProof(rootHash common.Hash, key []byte, proofNodes [][]byte) ([]byte, error) {
	wantHash := rootHash
	nibbles := keyToNibbles(key)

	for i, rlpNode := range proofNodes {
		nodeHash := crypto.Keccak256Hash(rlpNode)
		if len(rlpNode) >= 32 && wantHash != nodeHash {
			return nil, fmt.Errorf("invalid proof: hash mismatch at node %d", i)
		}

		node, err := trienode.DecodeNode(rlpNode)
		if err != nil {
			return nil, fmt.Errorf("failed to decode node: %v", err)
		}

		if err := node.Validate(nibbles); err != nil {
			return nil, fmt.Errorf("invalid node: %v", err)
		}

		switch n := node.(type) {
		case *trienode.LeafNode:
			return n.Value, nil
		case *trienode.ExtensionNode:
			nibbles = nibbles[len(n.Path):]
			wantHash = resolveHash(n.Next)
		case *trienode.BranchNode:
			index := nibbles[0]
			nibbles = nibbles[1:]
			child := n.Children[index]
			wantHash = resolveHash(child)
		default:
			return nil, fmt.Errorf("invalid node type")
		}
	}

	return nil, fmt.Errorf("incomplete proof")
}

// keyToNibbles converts a key byte slice to a
// slice of nibbles (half-bytes).
func keyToNibbles(key []byte) []byte {
	nibbles := make([]byte, len(key)*2)
	for i, b := range key {
		nibbles[i*2] = b >> 4
		nibbles[i*2+1] = b & 0x0F
	}

	return nibbles
}

// resolveHash returns the hash of the specified data.
func resolveHash(data []byte) common.Hash {
	if len(data) == common.HashLength {
		return common.Hash(data)
	}

	return crypto.Keccak256Hash(data)
}

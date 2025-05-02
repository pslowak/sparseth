package trienode

import (
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	// shortNodeLength is the length of either a
	// leaf node or extension node
	shortNodeLength = 2

	// fullNodeLength is the length of a branch node
	fullNodeLength = 17
)

// DecodeNode decodes a node from its RLP encoding.
func DecodeNode(rlpData []byte) (TrieNode, error) {
	var decoded []interface{}
	if err := rlp.DecodeBytes(rlpData, &decoded); err != nil {
		return nil, fmt.Errorf("RLP decode failed %v", err)
	}

	switch len(decoded) {
	case shortNodeLength:
		return decodeShortNode(decoded)
	case fullNodeLength:
		return decodeFullNode(decoded)
	default:
		return nil, fmt.Errorf("invalid node length %d", len(decoded))
	}
}

// decodeShortNode decodes a short node, i.e.,
// either a leaf node or a extension node.
func decodeShortNode(decoded []interface{}) (TrieNode, error) {
	compactPath, ok := decoded[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid short node path")
	}
	data, ok := decoded[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid short node data")
	}

	isLeaf, path := decodeCompactPath(compactPath)
	if isLeaf {
		return &LeafNode{
			Path:  path,
			Value: data,
		}, nil
	} else {
		return &ExtensionNode{
			Path: path,
			Next: data,
		}, nil
	}
}

// decodeFullNode decodes a full node, i.e., a branch node.
func decodeFullNode(decoded []interface{}) (TrieNode, error) {
	var children [fullNodeLength - 1][]byte
	for i := 0; i < fullNodeLength-1; i++ {
		if b, ok := decoded[i].([]byte); ok {
			children[i] = b
		} else {
			return nil, fmt.Errorf("invalid full node data")
		}
	}

	return &BranchNode{
		Children: children,
		Value:    decoded[fullNodeLength-1].([]byte),
	}, nil
}

// decodeCompactPath decodes a compact path to nibbles. A compact
// path is used for short nodes in the Merkle Patrica trie.
func decodeCompactPath(encodedCompactPath []byte) (bool, []byte) {
	if len(encodedCompactPath) == 0 {
		return false, nil
	}

	// Ethereum uses the following nibble encoding:
	// 0: extension node, even length
	// 1: extension node, odd length
	// 2: leaf node, even length
	// 3: leaf node, odd length
	typeAndParity := encodedCompactPath[0] >> 4
	isLeaf := (typeAndParity & 0x2) != 0
	oddLength := (typeAndParity & 0x1) != 0

	nibbles := make([]byte, 0)
	if oddLength {
		nibbles = append(nibbles, encodedCompactPath[0]&0xF)
	}

	for _, b := range encodedCompactPath[1:] {
		nibbles = append(nibbles, b>>4, b&0x0F)
	}

	return isLeaf, nibbles
}

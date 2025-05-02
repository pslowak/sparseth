package trienode

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

// TrieNode represents a node in a Merkle
// Patricia trie.
type TrieNode interface {
	// Validate validates whether this node is valid
	// for the given path. The specified path is
	// not modified.
	Validate(path []byte) error

	// String returns the string representation
	// of this node.
	String() string
}

// LeafNode represents a leaf node in a
// Merkle Patricia trie.
type LeafNode struct {
	// Path is the expanded path of the node.
	Path []byte

	// Value is the value of the node.
	Value []byte
}

func (l *LeafNode) Validate(path []byte) error {
	if !bytes.HasPrefix(path, l.Path) {
		return fmt.Errorf("path mismatch")
	}

	remaining := path[len(l.Path):]
	if len(remaining) != 0 {
		return fmt.Errorf("leaf node with remaining path")
	}

	return nil
}

func (l *LeafNode) String() string {
	path := hex.EncodeToString(l.Path)
	val := hex.EncodeToString(l.Value)

	return fmt.Sprintf("LeafNode{Path: %s, Value: %s}", path, val)
}

// ExtensionNode represents an extension node
// in a Merkle Patricia trie.
type ExtensionNode struct {
	// Path is the expanded path of the node.
	Path []byte

	// Next represents either an RLP encoded
	// node or a hash.
	Next []byte
}

func (e *ExtensionNode) Validate(path []byte) error {
	if !bytes.HasPrefix(path, e.Path) {
		return fmt.Errorf("path mismatch")
	}

	return nil
}

func (e *ExtensionNode) String() string {
	path := hex.EncodeToString(e.Path)
	next := hex.EncodeToString(e.Next)

	return fmt.Sprintf("ExtensionNode{Path: %s, Next: %s}", path, next)
}

// BranchNode represents a branch node in a
// Merkle Patricia trie.
type BranchNode struct {
	// Children are the children of the node.
	Children [fullNodeLength - 1][]byte

	// Value is the value of the node.
	// The value is non-nil if this node
	// terminates a key.
	Value []byte
}

func (b *BranchNode) String() string {
	val := "Empty"
	if len(b.Value) > 0 {
		val = hex.EncodeToString(b.Value)
	}

	var builder strings.Builder
	builder.WriteString("BranchNode{Children: [")
	for i, child := range b.Children {
		if len(child) > 0 {
			builder.WriteString(fmt.Sprintf("%d: %s, ", i, hex.EncodeToString(child)))
		}
	}
	builder.WriteString("], Value: ")
	builder.WriteString(val)
	builder.WriteString("}")

	return builder.String()
}

func (b *BranchNode) Validate(path []byte) error {
	if len(path) == 0 {
		return nil
	}

	index := path[0]
	if len(b.Children[index]) == 0 {
		return fmt.Errorf("missing branch at index %d", index)
	}

	return nil
}

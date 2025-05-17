package config

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"os"
)

// layout represents the top-level structure of
// a Solidity storage layout.
type layout struct {
	Storage []storageEntry       `json:"storage"`
	Types   map[string]typeEntry `json:"types"`
}

// storageEntry represents a single variable
// in the contract's storage.
type storageEntry struct {
	Label string `json:"label"`
	Type  string `json:"type"`
	Slot  string `json:"slot"`
}

// typeEntry represents the metadata for
// a Solidity type.
type typeEntry struct {
	Label string `json:"label"`
}

// LoadHeadSlot scans the storage layout file located at
// the specified path for a bytes32 variable named 'head'.
func LoadHeadSlot(path string) (common.Hash, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to read storage layout: %w", err)
	}

	var storageLayout layout
	if err = json.Unmarshal(data, &storageLayout); err != nil {
		return common.Hash{}, fmt.Errorf("failed to unmarshal storage layout: %w", err)
	}

	for _, entry := range storageLayout.Storage {
		if entry.Label == "head" && storageLayout.Types[entry.Type].Label == "bytes32" {
			return common.HexToHash("0x" + entry.Slot), nil
		}
	}

	return common.Hash{}, fmt.Errorf("no bytes32 field with label 'head' found in storage layout")
}

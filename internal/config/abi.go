package config

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"os"
	"strings"
)

// LoadABI reads an Ethereum smart contract ABI
// from the file at the specified path.
func LoadABI(path string) (abi.ABI, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	parsed, err := abi.JSON(strings.NewReader(string(data)))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return parsed, nil
}

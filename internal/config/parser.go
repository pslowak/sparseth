package config

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"os"
	"sparseth/internal/log"
	"strings"
)

// parser handles the conversion of raw config
// data into structured AccountsConfig data.
type parser struct {
	log log.Logger
}

// newParser creates a new parser
// with the specified logger.
func newParser(log log.Logger) *parser {
	return &parser{
		log: log.With("component", "config-parser"),
	}
}

// parse parses the raw config data
// into an AccountsConfig.
func (p *parser) parse(raw *config) (*AccountsConfig, error) {
	var accounts []*AccountConfig
	for _, unparsed := range raw.Accounts {
		parsed, err := p.parseAccount(unparsed)
		if err != nil {
			return nil, fmt.Errorf("failed to parse account: %w", err)
		}
		accounts = append(accounts, parsed)
	}

	return &AccountsConfig{
		Accounts: accounts,
	}, nil
}

// parseAccount parses a single account.
func (p *parser) parseAccount(acc *account) (*AccountConfig, error) {
	p.log.Debug("parse account", "address", acc.Address)

	addr := common.HexToAddress(acc.Address)

	if !acc.hasABI() {
		p.log.Debug("no ABI path detected", "address", addr.Hex())

		return &AccountConfig{
			Addr:           addr,
			ContractConfig: nil,
		}, nil
	}

	p.log.Debug("parse ABI", "address", addr.Hex(), "path", acc.ABI)
	contractABI, err := p.parseABI(acc.ABI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	headSlot, err := p.parseHead(acc.HeadSlot, acc.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event chain head: %w", err)
	}

	return &AccountConfig{
		Addr: addr,
		ContractConfig: &ContractConfig{
			ABI:      contractABI,
			HeadSlot: headSlot,
		},
	}, nil
}

// parseABI reads the ABI file and parses
// it into an Ethereum ABI structure.
func (p *parser) parseABI(path string) (abi.ABI, error) {
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

// parseHead resolves the head slot for the contract
// using the following precedence:
//  1. Explicit head slot
//  2. Storage layout file
//  3. Defaults to slot 0x0
func (p *parser) parseHead(slot string, path string) (common.Hash, error) {
	if slot != "" {
		p.log.Debug("head slot found in config", "slot", slot)
		return common.HexToHash(slot), nil
	}

	if path != "" {
		p.log.Debug("head slot not specified, fallback to storage layout", "path", path)
		head, err := parseHeadFromStorageLayout(path)
		if err == nil {
			return head, nil
		}
	}

	p.log.Debug("head slot not found, fallback to default value 0x0")
	return common.BigToHash(big.NewInt(0)), nil
}

// parseHeadFromStorageLayout scans the storage layout
// file located at the specified path for a bytes32
// variable named 'head'.
func parseHeadFromStorageLayout(path string) (common.Hash, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to read storage layout: %w", err)
	}

	type storageEntry struct {
		Label string `json:"label"`
		Type  string `json:"type"`
		Slot  string `json:"slot"`
	}

	type typeEntry struct {
		Label string `json:"label"`
	}

	type layout struct {
		Storage []storageEntry       `json:"storage"`
		Types   map[string]typeEntry `json:"types"`
	}

	var storageLayout layout
	if err = json.Unmarshal(data, &storageLayout); err != nil {
		return common.Hash{}, fmt.Errorf("failed to unmarshal storage layout: %w", err)
	}

	for _, entry := range storageLayout.Storage {
		if entry.Label == "head" && storageLayout.Types[entry.Type].Label == "bytes32" {
			slot := new(big.Int)
			if _, ok := slot.SetString(entry.Slot, 10); !ok {
				return common.Hash{}, fmt.Errorf("failed to parse slot: %s", entry.Slot)
			}
			return common.BigToHash(slot), nil
		}
	}

	return common.Hash{}, fmt.Errorf("no bytes32 field with label 'head' found in storage layout")
}

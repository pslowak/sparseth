package config

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"sparseth/log"
	"strings"
)

// empty is a constant used to
// represent the empty string.
var empty = ""

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

	p.log.Debug("parse event config", "address", addr.Hex())
	eventConfig, err := p.parseEventConfig(acc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event config: %w", err)
	}

	p.log.Debug("parse sparse config", "address", addr.Hex())
	sparseConfig, err := p.parseSparseConfig(acc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sparse config: %w", err)
	}

	return &AccountConfig{
		Addr: addr,
		ContractConfig: &ContractConfig{
			Event: eventConfig,
			State: sparseConfig,
		},
	}, nil
}

// parseEventConfig parses the event configuration
// for the specified account. Note that if no ABI
// is specified, and no head slot is found, this
// is no error and the returned EventConfig is nil.
func (p *parser) parseEventConfig(acc *account) (*EventConfig, error) {
	if acc.ABI == empty && acc.HeadSlot == empty {
		p.log.Debug("no event config found for account", "address", acc.Address)
		return nil, nil
	}

	head := common.HexToHash(acc.HeadSlot)
	contractAbi, err := p.parseABI(acc.ABI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI for account %s: %w", acc.Address, err)
	}

	return &EventConfig{
		ABI:      contractAbi,
		HeadSlot: head,
	}, nil
}

// parseSparseConfig parses the contract
// configuration for the specified account.
// Note that if no count slot is found, this
// is no error and the returned SparseConfig
// is nil.
func (p *parser) parseSparseConfig(acc *account) (*SparseConfig, error) {
	if acc.CountSlot == empty {
		p.log.Debug("no sparse contract config found for account", "address", acc.Address)
		return nil, nil
	}

	return &SparseConfig{
		CountSlot: common.HexToHash(acc.CountSlot),
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

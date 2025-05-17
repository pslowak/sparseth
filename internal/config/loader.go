package config

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v3"
	"math/big"
	"os"
	"sparseth/internal/log"
)

// AppConfig contains the top-level config
// structure for all Ethereum accounts to
// be monitored.
type AppConfig struct {
	Accounts []*AccountConfig
}

// AccountConfig holds the parsed config
// for an individual Ethereum account.
type AccountConfig struct {
	Addr common.Address
	Slot common.Hash
	ABI  abi.ABI
}

// config represents the raw YAML structure
// of the config file.
type config struct {
	Accounts []*account `yaml:"accounts"`
}

// account represents a raw YAML account entry.
type account struct {
	Address  string `yaml:"address"`
	ABI      string `yaml:"abi_path"`
	Storage  string `yaml:"storage_path"`
	HeadSlot string `yaml:"head_slot"`
}

// Loader reads the main config file.
type Loader struct {
	log log.Logger
}

// NewLoader creates a new config Loader with
// the specified logging context attached.
func NewLoader(log log.Logger) *Loader {
	return &Loader{
		log: log.With("component", "config-loader"),
	}
}

// Load reads the config file at the specified path.
func (l *Loader) Load(path string) (*AppConfig, error) {
	l.log.Info("load config")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var raw config
	if err = yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	var accounts []*AccountConfig
	for idx, unparsed := range raw.Accounts {
		if parsed, err := l.parseAccount(unparsed); err != nil {
			return nil, fmt.Errorf("failed to parse account at index %d: %w", idx, err)
		} else {
			accounts = append(accounts, parsed)
		}
	}

	return &AppConfig{
		Accounts: accounts,
	}, nil
}

// parseAccount transforms a raw YAML account
// into a structured AccountConfig.
func (l *Loader) parseAccount(acc *account) (*AccountConfig, error) {
	l.log.Debug("load account", "addr", acc.Address)

	if acc.Address == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if acc.ABI == "" {
		return nil, fmt.Errorf("account ABI is required")
	}

	parsedAddr := common.HexToAddress(acc.Address)
	parsedSlot, err := l.parseSlot(acc.HeadSlot, acc.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse slot: %w", err)
	}
	parsedABI, err := LoadABI(acc.ABI)
	if err != nil {
		return nil, fmt.Errorf("failed to load ABI: %w", err)
	}

	return &AccountConfig{
		Addr: parsedAddr,
		Slot: parsedSlot,
		ABI:  parsedABI,
	}, nil
}

// parseSlot resolves the storage slot
// to use for event tracking.
func (l *Loader) parseSlot(slot string, storage string) (common.Hash, error) {
	if slot != "" {
		return common.HexToHash(slot), nil
	} else if storage != "" {
		l.log.Debug("head not specified, fallback to storage layout")
		if head, err := LoadHeadSlot(storage); err == nil {
			return head, nil
		}
	}

	l.log.Debug("head not found in storage layout, fallback to 0x0")
	return common.BigToHash(big.NewInt(0)), nil
}

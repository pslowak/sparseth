package config

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// AccountsConfig contains the top-level
// config structure for all Ethereum
// accounts to be monitored.
type AccountsConfig struct {
	Accounts []*AccountConfig
}

// AccountConfig defines the monitoring
// params for a single Ethereum account.
type AccountConfig struct {
	// Addr is the address of the account.
	Addr common.Address
	// ContractConfig is nil for EOAs.
	ContractConfig *ContractConfig
}

// ContractConfig defines the monitoring
// params for a contract account.
type ContractConfig struct {
	// ABI defines the contract's application
	// binary interface.
	ABI abi.ABI
	// HeadSlot specifies the storage location
	// of the event hash chain head.
	HeadSlot common.Hash
	// CountSlot specifies the storage location
	// of the interaction counter.
	CountSlot common.Hash
}

// HasContractConfig checks if the account
// has a contract configuration, which is
// necessary for event monitoring.
func (a *AccountConfig) HasContractConfig() bool {
	return a.ContractConfig != nil
}

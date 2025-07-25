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
	// ContractConfig defines the monitoring
	// params for a contract account for both
	// event and state monitoring.
	ContractConfig *ContractConfig
}

// Contains checks whether the specified
// address is a monitored account.
func (a *AccountsConfig) Contains(addr common.Address) bool {
	for _, acc := range a.Accounts {
		if acc.Addr == addr {
			return true
		}
	}
	return false
}

// ContractConfig defines the monitoring
// params for a contract account.
type ContractConfig struct {
	Event *EventConfig
	State *SparseConfig
}

// EventConfig defines the monitoring params
// for a contract account's event monitoring.
type EventConfig struct {
	// ABI defines the contract's application
	// binary interface.
	ABI abi.ABI
	// HeadSlot specifies the storage location
	// of the event hash chain head.
	HeadSlot common.Hash
}

// SparseConfig defines the monitoring params
// for a contract account's state monitoring.
type SparseConfig struct {
	// CountSlot specifies the storage location
	// of the interaction counter.
	CountSlot common.Hash
}

// HasEventConfig checks if the account
// has an event configuration, which is
// necessary for event monitoring.
func (c *ContractConfig) HasEventConfig() bool {
	return c.Event != nil
}

// HasSparseConfig checks if the account
// has a sparse configuration, which is
// necessary for contract state monitoring,
// but not for EOA state monitoring.
func (c *ContractConfig) HasSparseConfig() bool {
	return c.State != nil
}

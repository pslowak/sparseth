package monitor

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// AccountInfo holds details about the
// monitored Ethereum account.
type AccountInfo struct {
	// Addr is the address of the account
	// to be monitored.
	Addr common.Address
	// ABI is the application binary interface
	// of the account to be monitored.
	ABI abi.ABI
	// Slot contains the head of the hash
	// chain.
	Slot common.Hash
	// InitialHead is the initial head
	// value of the event chain.
	InitialHead common.Hash
}

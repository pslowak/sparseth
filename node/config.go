package node

import (
	"github.com/ethereum/go-ethereum/params"
	"sparseth/internal/config"
)

// Config represents a collection of configuration
// values required to initialize and run the node.
type Config struct {
	// ChainConfig specifies the Ethereum
	// chain parameters to use.
	ChainConfig *params.ChainConfig
	// AccountsConfig contains the configuration
	// for all accounts to be monitored.
	AccsConfig *config.AccountsConfig
	// RpcURL specified the URL to use to connect
	// to the Ethereum RPC provider.
	RpcURL string
	// IsEventMode indicates whether the node
	// runs in event monitoring mode.
	IsEventMode bool
}

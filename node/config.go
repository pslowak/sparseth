package node

import (
	"github.com/ethereum/go-ethereum/params"
	"sparseth/internal/config"
)

// Config represents a collection of configuration
// values required to initialize and run the node.
type Config struct {
	ChainConfig *params.ChainConfig
	AppConfig   *config.AppConfig
	RpcURL      string
}

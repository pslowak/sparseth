package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// HeaderContext implements a minimal ChainContext,
// providing only the chain configuration
//
// This is used when executing transactions in
// isolation, without the need for a full consensus
// engine, see TxExecutor.ExecuteTxs.
type HeaderContext struct {
	Params *params.ChainConfig
}

// Engine returns the chain's consensus engine.
//
// Note that we return nil here, as we do not
// support consensus operations in this context.
func (hc *HeaderContext) Engine() consensus.Engine {
	return nil
}

// GetHeader retrieves a block header by its
// hash and number.
//
// Note that we return nil for any hash-number
// combination, as header lookup is not supported
// in this context.
func (hc *HeaderContext) GetHeader(_ common.Hash, _ uint64) *types.Header {
	return nil
}

// Config returns the chain's configuration.
func (hc *HeaderContext) Config() *params.ChainConfig {
	return hc.Params
}

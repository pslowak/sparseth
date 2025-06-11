package ethclient

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"math/big"
)

// txProvider provides verified
// transaction-related data via
// the Ethereum RPC API.
type txProvider struct {
	c *Client
}

// newTxProvider creates a new txProvider
// using the specified client.
func newTxProvider(client *Client) *txProvider {
	return &txProvider{
		c: client,
	}
}

// getTxsAtBlock retrieves and verifies all
// transactions at the specified block.
func (p *txProvider) getTxsAtBlock(ctx context.Context, header *types.Header) ([]*TransactionWithIndex, error) {
	txs, err := p.c.GetTransactionsAtBlock(ctx, header.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Verify completeness and integrity of the txs
	root := types.DeriveSha(txs, trie.NewStackTrie(nil))
	if root != header.TxHash {
		return nil, fmt.Errorf("transaction hash does not match block hash")
	}

	indexedTxs := make([]*TransactionWithIndex, len(txs))
	for i, tx := range txs {
		indexedTxs[i] = &TransactionWithIndex{
			Tx:    tx,
			Index: i,
		}
	}

	return indexedTxs, err
}

// createAccessList creates an access list for the
// specified transaction based on the state at the
// specified block number.
func (p *txProvider) createAccessList(ctx context.Context, tx *TransactionWithSender, blockNum *big.Int) (*types.AccessList, error) {
	return p.c.CreateAccessList(ctx, tx.Tx, tx.From, blockNum)
}

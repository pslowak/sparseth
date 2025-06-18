package sync

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"sparseth/ethstore"
	"sparseth/internal/log"
	"sparseth/storage"
)

// MockClient is a mock implementation of a
// consensus client. Later, the Altair Light
// Client Protocol will be used.
type MockClient struct {
	db  *ethstore.HeaderStore
	ec  *ethclient.Client
	log log.Logger
	pub chan<- *types.Header
}

// NewMockClient creates a new mock consensus
// client publishing new block headers at the
// returned channel.
func NewMockClient(log log.Logger, rpc *rpc.Client, db storage.KeyValStore) (*MockClient, <-chan *types.Header) {
	ch := make(chan *types.Header, 128)
	ec := ethclient.NewClient(rpc)
	store := ethstore.NewHeaderStore(db)

	return &MockClient{
		db:  store,
		ec:  ec,
		pub: ch,
		log: log.With("component", "sync-client"),
	}, ch
}

// RunContext starts the consensus client, i.e.,
// new block headers are fetched and published.
//
// Note that the mock client does not verify new
// block headers. Also, sync-up is very rudimentary,
// as it starts from the genesis block every time.
func (c *MockClient) RunContext(ctx context.Context) error {
	defer close(c.pub)

	latest, err := c.ec.HeaderByNumber(ctx, big.NewInt(int64(rpc.LatestBlockNumber)))
	if err != nil {
		return fmt.Errorf("failed to fetch latest block: %w", err)
	}
	c.log.Info("latest block", "num", latest.Number, "hash", latest.Hash().Hex())

	c.log.Info("start sync up")
	if err = c.syncUp(ctx, latest.Number.Uint64()); err != nil {
		return fmt.Errorf("failed to sync up: %w", err)
	}
	c.log.Info("sync up finished")

	return c.syncNew(ctx)
}

// syncUp fetches all block headers from
// the genesis block to the latest block.
func (c *MockClient) syncUp(ctx context.Context, latest uint64) error {
	// Start from genesis block
	genesis, err := c.ec.HeaderByNumber(ctx, big.NewInt(0))
	if err != nil {
		return fmt.Errorf("failed to fetch genesis block: %w", err)
	}
	if err = c.db.Put(genesis); err != nil {
		return fmt.Errorf("failed to store genesis block header: %w", err)
	}

	for num := uint64(1); num <= latest; num++ {
		c.log.Debug("download block header", "num", num)
		head, err := c.ec.HeaderByNumber(ctx, big.NewInt(int64(num)))
		if err != nil {
			return fmt.Errorf("failed to fetch header at block %d: %w", num, err)
		}
		if err = c.handleNewBlockHead(head); err != nil {
			c.log.Warn("failed to handle new block head", "num", head.Number, "hash", head.Hash().Hex(), "err", err)
			return err
		}
	}

	return nil
}

// syncNew listens for new block headers and
// publishes them to the execution layer.
func (c *MockClient) syncNew(ctx context.Context) error {
	c.log.Info("start new block sync")

	headers := make(chan *types.Header)

	sub, err := c.ec.SubscribeNewHead(ctx, headers)
	if err != nil {
		return fmt.Errorf("failed to subscribe to new head: %w", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case head := <-headers:
			if err = c.handleNewBlockHead(head); err != nil {
				c.log.Warn("failed to handle new block head", "hash", head.Hash().Hex(), "err", err)
			}
		case err = <-sub.Err():
			c.log.Error("subscription error", "err", err)
			return err
		case <-ctx.Done():
			c.log.Info("stop block sync")
			return nil
		}
	}
}

// handleNewBlockHead processes a new block header.
func (c *MockClient) handleNewBlockHead(head *types.Header) error {
	c.log.Info("block sync got new head", "hash", head.Hash())

	// Normally, we would verify the header here,
	// but for the mock client, we skip verification.
	if err := c.db.Put(head); err != nil {
		c.log.Error("failed to store new block header", "num", head.Number, "hash", head.Hash().Hex(), "err", err)
	}

	c.pub <- head
	return nil
}

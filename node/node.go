package node

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/sync/errgroup"
	"math/big"
	"sparseth/execution"
	"sparseth/execution/ethclient"
	"sparseth/execution/monitor"
	"sparseth/execution/monitor/event"
	"sparseth/execution/monitor/state"
	"sparseth/internal/config"
	"sparseth/internal/log"
	"sparseth/storage"
	"sparseth/storage/mem"
	"sparseth/sync"
)

// Node is the coordinator of the node's
// various subsystems, such as the consensus
// client, block listener and monitors.
type Node struct {
	config *Config
	disp   *execution.Dispatcher
	db     storage.KeyValStore
	rpc    *rpc.Client
	log    log.Logger
}

// NewNode initializes a new Node instance
// with the provided configuration.
func NewNode(ctx context.Context, config *Config, log log.Logger) (*Node, error) {
	conn, err := rpc.DialContext(ctx, config.RpcURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to RPC provider: %w", err)
	}

	// Use an in-memory db (for now)
	db := mem.New()
	disp := execution.NewDispatcher(log)

	return &Node{
		config: config,
		disp:   disp,
		db:     db,
		rpc:    conn,
		log:    log.With("component", "node"),
	}, nil
}

// Start launches the consensus and
// execution clients of the node.
func (n *Node) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	consensus, pipe := sync.NewMockClient(n.log, n.rpc, n.db)
	listener := execution.NewListener(pipe, n.disp, n.log)
	ec := ethclient.NewClient(n.rpc)

	if n.config.IsEventMode {
		// Start up a single log monitor for each contract account
		for _, acc := range n.config.AccsConfig.Accounts {
			if acc.ContractConfig.HasEventConfig() {
				n.log.Info("start event monitor", "account", acc.Addr.Hex())
				g.Go(n.startEventMonitor(ctx, ec, acc))
			}
		}
	} else {
		// Start up a single transaction monitor for all accounts
		n.log.Info("start transaction monitor")
		g.Go(n.startTxMonitor(ctx, ec))
	}

	n.log.Info("start block listener")
	g.Go(n.startBlockListener(ctx, listener))

	n.log.Info("start consensus client")
	g.Go(n.startConsensusClient(ctx, consensus))

	if err := g.Wait(); err != nil {
		n.log.Error("failed to start node", "err", err)
		return fmt.Errorf("failed to start node: %w", err)
	}

	return nil
}

// Shutdown gracefully stops the node.
func (n *Node) Shutdown() {
	n.log.Info("shut down")

	n.rpc.Close()
	n.disp.Close()
	n.db.Close()
}

// startTxMonitor initializes and runs a transaction monitor.
func (n *Node) startTxMonitor(ctx context.Context, ec *ethclient.Client) func() error {
	return func() error {
		sub := n.disp.Subscribe("transaction-monitor")
		proc := state.NewTxProcessor(n.config.AccsConfig, n.config.ChainConfig, n.db, ec, n.log)
		mntr := monitor.NewMonitor("transaction", sub, proc, n.log)

		if err := mntr.RunContext(ctx); err != nil {
			n.log.Error("failed to start transaction-monitor", "err", err)
			return fmt.Errorf("failed to start transaction-monitor: %w", err)
		}

		return nil
	}
}

// startEventMonitor initializes and runs an event monitor
// for a specific account.
func (n *Node) startEventMonitor(ctx context.Context, ec *ethclient.Client, acc *config.AccountConfig) func() error {
	return func() error {
		info := &monitor.AccountInfo{
			Addr:        acc.Addr,
			ABI:         acc.ContractConfig.Event.ABI,
			Slot:        acc.ContractConfig.Event.HeadSlot,
			InitialHead: common.BigToHash(big.NewInt(0)),
		}

		sub := n.disp.Subscribe(acc.Addr.Hex())
		proc := event.NewLogProcessor(info, ec, n.db, n.log)
		mntr := monitor.NewMonitor(acc.Addr.Hex()+"-event", sub, proc, n.log)

		if err := mntr.RunContext(ctx); err != nil {
			n.log.Error("failed to start event-monitor", "err", err, "account", acc.Addr.Hex())
			return fmt.Errorf("failed to start event-monitor for %s: %w", acc.Addr.Hex(), err)
		}

		return nil
	}
}

// startBlockListener runs the block listener.
func (n *Node) startBlockListener(ctx context.Context, l *execution.Listener) func() error {
	return func() error {
		if err := l.RunContext(ctx); err != nil {
			n.log.Error("failed to start block listener", "err", err)
			return fmt.Errorf("failed to start block listener: %w", err)
		}
		return nil
	}
}

// startConsensusClient runs the consensus client.
func (n *Node) startConsensusClient(ctx context.Context, c *sync.MockClient) func() error {
	return func() error {
		if err := c.RunContext(ctx); err != nil {
			n.log.Error("failed to start block listener", "err", err)
			return fmt.Errorf("failed to start block listener: %w", err)
		}
		return nil
	}
}

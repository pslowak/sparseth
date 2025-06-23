package node

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/sync/errgroup"
	"sparseth/execution"
	"sparseth/execution/ethclient"
	"sparseth/execution/monitor"
	"sparseth/execution/monitor/state"
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

	n.log.Info("start transaction monitor")
	ec := ethclient.NewClient(n.rpc)
	sub := n.disp.Subscribe("transaction-monitor")
	proc := state.NewTxProcessor(n.config.ChainConfig, n.db, ec, n.log)
	mntr := monitor.NewMonitor("transaction", sub, proc, n.log)
	g.Go(func() error {
		if err := mntr.RunContext(ctx); err != nil {
			n.log.Error("failed to start state-monitor", "err", err)
			return fmt.Errorf("failed to start state-monitor: %w", err)
		}
		return nil
	})

	n.log.Info("start consensus client")
	g.Go(func() error {
		if err := consensus.RunContext(ctx); err != nil {
			n.log.Error("failed to start consensus client", "err", err)
			return fmt.Errorf("failed to start consensus client: %w", err)
		}
		return nil
	})

	n.log.Info("start block listener")
	g.Go(func() error {
		if err := listener.RunContext(ctx); err != nil {
			n.log.Error("failed to start block listener", "err", err)
			return fmt.Errorf("failed to start block listener: %w", err)
		}
		return nil
	})

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

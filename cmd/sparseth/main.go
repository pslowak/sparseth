package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"sparseth/internal/config"
	"sparseth/internal/log"
	"sparseth/node"
	"syscall"
)

func main() {
	rpcURL := flag.String("rpc", "ws://localhost:8545", "RPC provider URL to connect to")
	configPath := flag.String("config", "config.yaml", "Path to config file")
	eventModeFlag := flag.Bool("event-mode", false, "Enable event monitoring mode (default: false)")
	flag.Parse()

	logger := log.New(log.NewTerminalHandler()).With("component", "main")
	logger.Info("using RPC provider", "url", *rpcURL)
	logger.Info("using config file", "path", *configPath)
	logger.Info("event mode", "enabled", *eventModeFlag)

	loader := config.NewLoader(logger)
	accsConfig, err := loader.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	nodeConfig := &node.Config{
		ChainConfig: config.AnvilChainConfig, // The only chain supported (for now)
		AccsConfig:  accsConfig,
		RpcURL:      *rpcURL,
		IsEventMode: *eventModeFlag,
	}

	n, err := node.NewNode(ctx, nodeConfig, logger)
	if err != nil {
		logger.Error("failed to create node", "err", err)
		os.Exit(1)
	}
	defer n.Shutdown()

	logger.Info("start node")
	go func() {
		if err = n.Start(ctx); err != nil {
			logger.Error("node run failed", "err", err)
			cancel()
		}
	}()

	<-ctx.Done()

	if ctx.Err() != nil && !errors.Is(ctx.Err(), context.Canceled) {
		logger.Error("shutdown due to error", "err", ctx.Err())
		os.Exit(1)
	}

	logger.Info("graceful shutdown")
}

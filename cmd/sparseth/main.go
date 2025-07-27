package main

import (
	"context"
	"errors"
	"flag"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"os"
	"os/signal"
	userconfig "sparseth/config"
	internalconfig "sparseth/internal/config"
	"sparseth/internal/log"
	"sparseth/node"
	"syscall"
)

func main() {
	rpcURL := flag.String("rpc", "ws://localhost:8545", "RPC provider URL to connect to")
	configPath := flag.String("config", "config.yaml", "Path to config file")
	networkFlag := flag.String("network", "mainnet", "Ethereum network to use")
	eventModeFlag := flag.Bool("event-mode", false, "Enable event monitoring mode (default: false)")
	checkPointFlag := flag.String("checkpoint", "", "Checkpoint hash to start from (default: genesis hash of the network)")

	if v := os.Getenv("EXECUTION_RPC_URL"); v != "" {
		flag.Set("rpc", v)
	}
	if v := os.Getenv("CONFIG_PATH"); v != "" {
		flag.Set("config", v)
	}
	if v := os.Getenv("ETHEREUM_NETWORK"); v != "" {
		flag.Set("network", v)
	}
	if v := os.Getenv("CHECKPOINT_HASH"); v != "" {
		flag.Set("checkpoint", v)
	}
	if v := os.Getenv("EVENT_MODE"); v == "1" || v == "true" {
		flag.Set("event-mode", "true")
	}

	flag.Parse()

	logger := log.New(log.NewTerminalHandler()).With("component", "main")

	supportedNetworks := map[string]*params.ChainConfig{
		"mainnet": userconfig.MainnetChainConfig,
		"sepolia": userconfig.SepoliaChainConfig,
		"anvil":   userconfig.AnvilChainConfig,
	}

	chainConfig, exists := supportedNetworks[*networkFlag]
	if !exists {
		logger.Error("unsupported network", "network", *networkFlag)
		logger.Info("supported networks: mainnet, sepolia, anvil")
		os.Exit(2)
	}

	checkpoint := common.HexToHash(*checkPointFlag)
	if *checkPointFlag == "" {
		checkpoints := map[string]common.Hash{
			"mainnet": userconfig.MainnetGenesisHash,
			"sepolia": userconfig.SepoliaGenesisHash,
			"anvil":   userconfig.AnvilGenesisHash,
		}
		checkpoint = checkpoints[*networkFlag]
	}

	logger.Info("using RPC provider", "url", *rpcURL)
	logger.Info("using network", "name", *networkFlag)
	logger.Info("using checkpoint", "hash", checkpoint.Hex())
	logger.Info("using config file", "path", *configPath)
	logger.Info("event mode", "enabled", *eventModeFlag)

	loader := internalconfig.NewLoader(logger)
	accsConfig, err := loader.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	nodeConfig := &node.Config{
		ChainConfig: chainConfig,
		Checkpoint:  checkpoint,
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

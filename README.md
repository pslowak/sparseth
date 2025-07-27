# SPARSETH — A Sparse Node Protocol for Ethereum

![test](https://github.com/pslowak/sparseth/actions/workflows/go-test.yml/badge.svg)

SPARSETH is a lightweight sparse node protocol for Ethereum written in Go.

## Quick Start

Run the sparse node:

```bash
sparseth --rpc <ETHEREUM_RPC_URL>
```

## Requirements

Before building the node, make sure you have the following tools installed:
- Go 1.24+
- `solc` (Solidity compiler)
- `anvil` (local Ethereum development node)

Once installed, you can build the node using:

```bash
make all
```

If you prefer using Docker, you can build and run the node with:

```bash
docker compose up
```

## Usage

SPARSETH supports a variety of command-line options to configure its behavior:

```bash
sparseth [--rpc <url>] [--config <path>] [--network <name>] [--checkpoint <hash>] [--event-mode]
```

### Options

`--rpc <url>` URL of the Ethereum RPC endpoint to connect to (default: `ws://localhost:8545`).

`--config <path>` Path to the configuration file defining all monitored accounts (default: `config.yaml`).

`--network <name>` Name of the Ethereum network to connect to (default: `mainnet`). Supported networks are: `mainnet`,
`sepolia`, and `anvil`.

`--checkpoint <hash>` Hash of the block to start syncing from (default: `genesis` of the selected network). Note: You 
must explicitly provide this if you're running an Anvil node with non-default options. Your contract should be deployed
_after_ the specified checkpoint block.

`--event-mode` Enables _event mode_. If omitted, the node operates in _sparse mode_.


## Node Modes

SPARSETH supports two modes of operation:
- _Event mode_ – monitors events emitted by Ethereum smart contracts
- _Sparse mode_ – monitors the state of Ethereum accounts

> See the [Smart Contract Compatibility Guide](https://github.com/pslowak/sparseth/wiki/Smart-Contract-Compatibility-Guide)
to learn how to make your smart contract compatible with SPARSETH's execution modes.
 
### Event Mode

In event mode, the node listens for events emitted specific smart contracts. To make these events verifiable, each
contract must implement a hash chain that links each emitted event to the previous one using a cryptographic hash 
function:

$$H_n = \mathrm{hash}(H_{n-1}||\mathrm{event}_n)$$

Here, $H_n$ is the current hash chain head, $H_{n-1}$ is the previous head, and $\mathrm{event}_n$ represents the 
contents of the current event. The latest hash chain head is stored within the contract. Upon receiving new events, the
node computes the new hash chain head and compares it with the one stored in the contract, thereby ensuring that the
received events are not tampered with (integrity) or selectively omitted (completeness).

### Sparse Mode

In sparse mode, the node monitors the state of specific Ethereum accounts by maintaining a _sparse state_ (a minimal
subset of the global state that is relevant only to the monitored accounts). The node reconstructs this partial state by 
re-executing transactions that affect the monitored accounts. 

Monitoring contract accounts in sparse mode introduces the challenge of _transaction completeness_: ensuring that all
transactions affecting the contract's state have been received and processed by the node. To address this, the node 
requires each monitored contract to maintain a monotonic integer counter, which is incremented on every state-changing 
function call. By comparing the current value of the counter on-chain with the value of the counter reconstructed 
through local re-execution, the node can verify transaction completeness.

> Note: This approach would be most effective with support for transaction inclusion proofs. With such proofs, the node
could avoid downloading all transactions in a block and reconstructing the entire transaction trie. Instead, it could
fetch only the relevant transactions and verify their inclusion. However, such proofs are currently not available via 
the standard Ethereum RPC API.

## Configuration

SPARSETH uses a `config.yaml` file to define monitored accounts. For a quick overview, see the example below.

### Example Configuration

```yaml
accounts:
  - address: "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef" # required
    abi_path: "path/to/abi" # required in event mode
    head_slot: "0x0" # required in event mode
    count_slot: "0x1" # required in sparse mode for contract monitoring
```

> For detailed configuration options, refer to the [Configuration Guide](https://github.com/pslowak/sparseth/wiki/Configuration-Guide).

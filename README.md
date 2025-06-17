# SPARSETH — A Sparse Node Protocol for Ethereum

![test](https://github.com/pslowak/sparseth/actions/workflows/go-test.yml/badge.svg)

SPARSETH is a lightweight sparse node protocol for Ethereum written in Go.

## Quick Start

Run the sparse node:

```bash
sparseth --rpc "<ETHEREUM_RPC_URL>"
```

## Node Modes

SPARSETH supports two modes of operation:
- _Event mode_ – monitors events emitted by Ethereum smart contracts
- _Sparse mode_ – monitors the state of Ethereum accounts

### Event Mode

In event mode, the node listens for events emitted specific smart contracts. To make these events verifiable, each
contract must implement a hash chain that links each emitted event to the previous one using a cryptographic hash 
function:

$$H_n = \mathrm{hash}(H_{n-1}||\mathrm{event}_n)$$

Here, $H_n$ is the current hash chain head, $H_{n-1}$ is the previous head, and $\mathrm{event}_n$ represents the 
contents of the current event. The latest hash chain head is stored within the contract. Upon receiving new events, the
node computes the new hash chain head and compares it with the one stored in the contract, thereby ensuring that the
received events are not tampered with (integrity) or selectively omitted (completeness).

> See the [Smart Contract Compatibility Guide](https://github.com/pslowak/sparseth/wiki/Smart-Contract-Compatibility-Guide) 
to learn how to make your smart contract compatible with SPARSETH's event mode.

### Sparse Mode

In sparse mode, the node monitors the state of specific Ethereum accounts by maintaining a _sparse state_ (a minimal
subset of the global state that is relevant only to the monitored accounts). The node reconstructs this partial state by 
re-executing transactions that affect the monitored accounts. 

## Configuration

SPARSETH uses a `config.yaml` file to define monitored accounts.

### Example Configuration

```yaml
accounts:
  - address: "0x0000000000000000000000000000000000000000" # required
    abi_path: "path/to/abi" # required (event mode)
    head_slot: "0x0" # optional
    storage_path: "path/to/storage/layout" # optional
```

### Slot Resolution Priority

When determining the storage slot for the hash chain head, SPARSETH follows this order:

1. `head_slot` – Explicitly defined storage slot (e.g., `0x0`).
2. `storage_path` – Auto-detects a `bytes32` variable named `head`in the storage layout.
3. Default – Falls back to storage slot `0x0`.

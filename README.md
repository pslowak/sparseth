# SPARSETH — A Sparse Node Protocol for Ethereum

![test](https://github.com/pslowak/sparseth/actions/workflows/go-test.yml/badge.svg)

SPARSETH is a lightweight sparse node protocol for Ethereum written in Go.

## Quick Start

Run the sparse node:

```bash
sparseth --rpc "<ETHEREUM_RPC_URL>"
```

## Configuration

SPARSETH uses a `config.yaml` file to define monitored contracts.

### Example Configuration

```yaml
accounts:
  - address: "0x0000000000000000000000000000000000000000" # required
    abi_path: "path/to/abi" # required
    head_slot: "0x0" # optional
    storage_path: "path/to/storage/layout" # optional
```

### Slot Resolution Priority

When determining the storage slot for the hash chain head, SPARSETH follows this order:

1. `head_slot` – Explicitly defined storage slot (e.g., `0x0`).
2. `storage_path` – Auto-detects a `bytes32` variable named `head`in the storage layout.
3. Default – Falls back to storage slot `0x0`.

## Smart Contract Compatability

For SPARSETH to track events efficiently, your contract must implement a hash chain
mechanism. See `Storage.sol` for a reference implementation.

### Requirements

- All events tracked in the hash chain must be non-anonymous.
- The contract must maintain a `bytes32` variable updated with each tracked event.
- Event fields must be hashed in the declared order.

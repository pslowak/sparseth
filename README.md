# SPARSETH—A Sparse Node Protocol for Ethereum

Sparse node protocol for Ethereum written in Go.

![test](https://github.com/pslowak/sparseth/actions/workflows/go-test.yml/badge.svg)

## Usage

### Event Tracking

SPARSETH is designed to efficiently track smart contract events using a hash chain mechanism.
To make a contract compatible with the sparse node, follow the approach shown in `Storage.sol`.
The contract maintains a `head` value which is updated on each emitted event by hashing the 
previous `head` together with all event fields in the declared order.

By default, the sparse node looks for a `bytes32` variable named `head` in the contract's 
storage layout. If such a variable is not found, the sparse node falls back to reading 
from storage slot index `0`.

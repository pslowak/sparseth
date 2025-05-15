package event

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Verifier verifies the completeness and integrity
// of Ethereum event logs using a hash chain mechanism.
//
// All events must be non-anonymous.
type Verifier struct {
	// abi is the ABI of the contract.
	abi abi.ABI
	// head is the current head of the hash chain.
	head common.Hash
}

// NewLogVerifier creates a new Verifier
// instance for the specified contract ABI
// and initial hash chain head.
//
// The abi must include all definitions of
// all events that will be verified.
func NewLogVerifier(abi abi.ABI, head common.Hash) *Verifier {
	return &Verifier{
		abi:  abi,
		head: head,
	}
}

// VerifyLogs validates the specified ordered slice
// of logs against the expected hash chain head.
func (v *Verifier) VerifyLogs(logs []types.Log, expected common.Hash) error {
	curr := v.head

	for _, l := range logs {
		var err error
		if curr, err = v.computeNewHead(curr, l); err != nil {
			return fmt.Errorf("failed to compute new event head: %w", err)
		}
	}

	if !bytes.Equal(curr.Bytes(), expected.Bytes()) {
		return fmt.Errorf("head mismatch")
	}

	v.head = curr
	return nil
}

// computeNewHead calculates the new hash chain
// head after processing a single log.
func (v *Verifier) computeNewHead(prev common.Hash, log types.Log) (common.Hash, error) {
	if len(log.Topics) < 1 {
		return common.Hash{}, fmt.Errorf("log does not contain ID")
	}

	id := log.Topics[0]
	event, err := v.abi.EventByID(id)
	if err != nil {
		return common.Hash{}, fmt.Errorf("unknown event ID: %w", err)
	}

	data, err := event.Inputs.NonIndexed().UnpackValues(log.Data)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to unpack log data: %w", err)
	}

	if len(event.Inputs.NonIndexed()) != len(data) {
		return common.Hash{}, fmt.Errorf("non-indexed event data mismatch: want %d, got %d", len(event.Inputs.NonIndexed()), len(data))
	}

	// Prepend the head
	args := abi.Arguments{
		abi.Argument{
			Name: "head",
			Type: abi.Type{
				T:    abi.FixedBytesTy,
				Size: 32,
			},
		},
	}
	vals := []interface{}{prev}

	indexed, nonIndexed := 1, 0
	for _, arg := range event.Inputs {
		args = append(args, arg)
		if arg.Indexed {
			if len(log.Topics) <= indexed {
				return common.Hash{}, fmt.Errorf("topic count mismatch: want %d, got %d", indexed, len(log.Topics)-1)
			}
			vals = append(vals, log.Topics[indexed])
			indexed++
		} else {
			vals = append(vals, data[nonIndexed])
			nonIndexed++
		}
	}

	if indexed != len(log.Topics) {
		topics := len(event.Inputs) - len(event.Inputs.NonIndexed())
		return common.Hash{}, fmt.Errorf("topic count mismatch: want %d, got %d", topics, len(log.Topics)-1)
	}

	packed, err := args.Pack(vals...)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack args: %w", err)
	}

	return crypto.Keccak256Hash(packed), nil
}

package event

import (
	"bytes"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"testing"
)

func TestVerifier_VerifyLogs(t *testing.T) {
	erc20abi, err := abi.JSON(bytes.NewReader([]byte("[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]")))
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	t.Run("should return error when abi does not match (missing field)", func(t *testing.T) {
		transferEvent := erc20abi.Events["Transfer"]
		data, err := transferEvent.Inputs.NonIndexed().Pack(big.NewInt(1))
		if err != nil {
			t.Fatalf("failed to pack event: %v", err)
		}

		logs := []*types.Log{
			{
				Topics: []common.Hash{
					transferEvent.ID, // Signature
					common.BigToHash(common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266").Big()), // 'from'
					// Missing 'to'
				},
				Data: data,
			},
		}

		verifier := NewLogVerifier(erc20abi, common.Hash{})
		if err = verifier.VerifyLogs(logs, common.Hash{}); err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should return error when abi does not match (additional field)", func(t *testing.T) {
		transferEvent := erc20abi.Events["Transfer"]
		data, err := transferEvent.Inputs.NonIndexed().Pack(big.NewInt(1))
		if err != nil {
			t.Fatalf("failed to pack event: %v", err)
		}

		logs := []*types.Log{
			{
				Topics: []common.Hash{
					transferEvent.ID, // Signature
					common.BigToHash(common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266").Big()), // 'from'
					common.BigToHash(common.HexToAddress("0xa513e6e4b8f2a923d98304ec87f64353c4d5c853").Big()), // 'to'
					common.BigToHash(common.HexToAddress("0xabcd12345678a923d98304ec87f64353c4d5c853").Big()), // additional
				},
				Data: data,
			},
		}

		verifier := NewLogVerifier(erc20abi, common.Hash{})
		if err = verifier.VerifyLogs(logs, common.Hash{}); err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should return error when log is missing", func(t *testing.T) {
		transferEvent := erc20abi.Events["Transfer"]
		transferData, err := transferEvent.Inputs.NonIndexed().Pack(big.NewInt(1))
		if err != nil {
			t.Fatalf("failed to pack event: %v", err)
		}

		// Emitted 'transfer' and 'approval', but got only 'transfer'
		logs := []*types.Log{
			{
				Topics: []common.Hash{
					transferEvent.ID,
					common.BigToHash(common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266").Big()),
					common.BigToHash(common.HexToAddress("0xa513e6e4b8f2a923d98304ec87f64353c4d5c853").Big()),
				},
				Data: transferData,
			},
		}

		current := common.HexToHash("0xfe64ba9e577c4903954c702589370173f0849780586a5ff634e0faf0bdc24db9")
		expected := common.HexToHash("0x969902f40d276b80ebebe0ff50f874203b0adc522c34f9266cc487cc59b94e76")
		verifier := NewLogVerifier(erc20abi, current)
		if err = verifier.VerifyLogs(logs, expected); err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should return error when log is corrupted", func(t *testing.T) {
		transferEvent := erc20abi.Events["Transfer"]
		transferData, err := transferEvent.Inputs.NonIndexed().Pack(big.NewInt(1))
		if err != nil {
			t.Fatalf("failed to pack event: %v", err)
		}
		approvalEvent := erc20abi.Events["Approval"]
		approvalData, err := approvalEvent.Inputs.NonIndexed().Pack(big.NewInt(-2)) // value corrupted
		if err != nil {
			t.Fatalf("failed to pack event: %v", err)
		}

		logs := []*types.Log{
			{
				Topics: []common.Hash{
					transferEvent.ID,
					common.BigToHash(common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266").Big()),
					common.BigToHash(common.HexToAddress("0xa513e6e4b8f2a923d98304ec87f64353c4d5c853").Big()),
				},
				Data: transferData,
			},
			{
				Topics: []common.Hash{
					approvalEvent.ID,
					common.BigToHash(common.HexToAddress("0xa513e6e4b8f2a923d98304ec87f64353c4d5c853").Big()),
					common.BigToHash(common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266").Big()),
				},
				Data: approvalData,
			},
		}

		current := common.HexToHash("0xfe64ba9e577c4903954c702589370173f0849780586a5ff634e0faf0bdc24db9")
		expected := common.HexToHash("0x969902f40d276b80ebebe0ff50f874203b0adc522c34f9266cc487cc59b94e76")
		verifier := NewLogVerifier(erc20abi, current)
		if err = verifier.VerifyLogs(logs, expected); err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should verify correct logs without error", func(t *testing.T) {
		transferEvent := erc20abi.Events["Transfer"]
		transferData, err := transferEvent.Inputs.NonIndexed().Pack(big.NewInt(1))
		if err != nil {
			t.Fatalf("failed to pack event: %v", err)
		}
		approvalEvent := erc20abi.Events["Approval"]
		approvalData, err := approvalEvent.Inputs.NonIndexed().Pack(big.NewInt(2))
		if err != nil {
			t.Fatalf("failed to pack event: %v", err)
		}

		logs := []*types.Log{
			{
				Topics: []common.Hash{
					transferEvent.ID,
					common.BigToHash(common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266").Big()),
					common.BigToHash(common.HexToAddress("0xa513e6e4b8f2a923d98304ec87f64353c4d5c853").Big()),
				},
				Data: transferData,
			},
			{
				Topics: []common.Hash{
					approvalEvent.ID,
					common.BigToHash(common.HexToAddress("0xa513e6e4b8f2a923d98304ec87f64353c4d5c853").Big()),
					common.BigToHash(common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266").Big()),
				},
				Data: approvalData,
			},
		}

		current := common.HexToHash("0xfe64ba9e577c4903954c702589370173f0849780586a5ff634e0faf0bdc24db9")
		expected := common.HexToHash("0x969902f40d276b80ebebe0ff50f874203b0adc522c34f9266cc487cc59b94e76")
		verifier := NewLogVerifier(erc20abi, current)
		if err = verifier.VerifyLogs(logs, expected); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should not update head on error", func(t *testing.T) {
		transferEvent := erc20abi.Events["Transfer"]
		transferData, err := transferEvent.Inputs.NonIndexed().Pack(big.NewInt(1))
		if err != nil {
			t.Fatalf("failed to pack event: %v", err)
		}

		// Emitted 'transfer' and 'approval', but got only 'transfer'
		logs := []*types.Log{
			{
				Topics: []common.Hash{
					transferEvent.ID,
					common.BigToHash(common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266").Big()),
					common.BigToHash(common.HexToAddress("0xa513e6e4b8f2a923d98304ec87f64353c4d5c853").Big()),
				},
				Data: transferData,
			},
		}

		current := common.HexToHash("0xfe64ba9e577c4903954c702589370173f0849780586a5ff634e0faf0bdc24db9")
		expected := common.HexToHash("0x969902f40d276b80ebebe0ff50f874203b0adc522c34f9266cc487cc59b94e76")
		verifier := NewLogVerifier(erc20abi, current)
		if err = verifier.VerifyLogs(logs, expected); err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !bytes.Equal(current.Bytes(), verifier.head.Bytes()) {
			t.Errorf("expected head to be unchanged to %s, got %s", current.Bytes(), verifier.head.Bytes())
		}
	})

	t.Run("should update head on success", func(t *testing.T) {
		transferEvent := erc20abi.Events["Transfer"]
		transferData, err := transferEvent.Inputs.NonIndexed().Pack(big.NewInt(1))
		if err != nil {
			t.Fatalf("failed to pack event: %v", err)
		}
		approvalEvent := erc20abi.Events["Approval"]
		approvalData, err := approvalEvent.Inputs.NonIndexed().Pack(big.NewInt(2))
		if err != nil {
			t.Fatalf("failed to pack event: %v", err)
		}

		logs := []*types.Log{
			{
				Topics: []common.Hash{
					transferEvent.ID,
					common.BigToHash(common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266").Big()),
					common.BigToHash(common.HexToAddress("0xa513e6e4b8f2a923d98304ec87f64353c4d5c853").Big()),
				},
				Data: transferData,
			},
			{
				Topics: []common.Hash{
					approvalEvent.ID,
					common.BigToHash(common.HexToAddress("0xa513e6e4b8f2a923d98304ec87f64353c4d5c853").Big()),
					common.BigToHash(common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266").Big()),
				},
				Data: approvalData,
			},
		}

		current := common.HexToHash("0xfe64ba9e577c4903954c702589370173f0849780586a5ff634e0faf0bdc24db9")
		expected := common.HexToHash("0x969902f40d276b80ebebe0ff50f874203b0adc522c34f9266cc487cc59b94e76")
		verifier := NewLogVerifier(erc20abi, current)
		if err = verifier.VerifyLogs(logs, expected); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !bytes.Equal(verifier.head.Bytes(), expected.Bytes()) {
			t.Errorf("expected head to be updated to %s, got %s", expected.Bytes(), verifier.head.Bytes())
		}
	})
}

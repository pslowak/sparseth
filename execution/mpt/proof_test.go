package mpt

import (
	"bytes"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"strings"
	"testing"
)

func TestVerifyAccountProof(t *testing.T) {
	t.Run("should verify valid account proof", func(t *testing.T) {
		stateRoot := common.HexToHash("0x0136b96aa9d793cdccd5d1f4f03a576b0f64ce562dcb8d423414b5cff37e3d6c")
		address := common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266")
		proof := []string{
			"0xf90131a0b91a8b7a7e9d3eab90afd81da3725030742f663c6ed8c26657bf00d842a9f4aaa01689b2a5203afd9ea0a0ca3765e4a538c7176e53eac1f8307a344ffc3c6176558080a0de673157fb5e8d14d783c948b64074922bf60224389cb46a3d38d48a7e81ae4ea04d5794121ef1a51608fa5b655bb3f861fb0a4fcecf8b7fecbf084b2d422a8bcf8080a04b29efa44ecf50c19b34950cf1d0f05e00568bcc873120fbea9a4e8439de0962a0d0a1bfe5b45d2d863a794f016450a4caca04f3b599e8d1652afca8b752935fd880a0bf9b09e442e044778b354abbadb5ec049d7f5e8b585c3966d476c4fbc9a181d28080a0a3a8f2834a8836fa2e4824f6c1dbe936a895fcfd53965acdf896567b138b90f6a0e5c557a0ce3894afeb44c37f3d24247f67dc76a174d8cacc360c1210eef60a7680",
			"0xf8518080808080a0aabfb1441169c3379f428df147ba34658049e31ab75bca31dcea5ea3513408a7808080a0df27128ae81e00b9ab17d7c0ff1fe52aa0320efba06361a8d6e9934daa27e76080808080808080",
			"0xf873a020707d0e6171f728f7473c24cc0432a9b07eaaf1efed6a137a4a8c12c79552d9b850f84e018a021e19e053fa587ede00a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a0c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
		}

		proofNodes := make([][]byte, len(proof))
		for idx, node := range proof {
			bytez, err := hex.DecodeString(strings.TrimPrefix(node, "0x"))
			if err != nil {
				t.Fatalf("failed to decode node %d %v", idx, node)
			}

			proofNodes[idx] = bytez
		}

		account, err := VerifyAccountProof(stateRoot, address, proofNodes)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		expectedNonce := uint64(1)
		expectedBalance := new(big.Int)
		expectedBalance.SetString("21e19e053fa587ede00", 16)
		expectedCodeHash := common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
		expectedStorageRoot := common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

		if account.Nonce != expectedNonce {
			t.Errorf("expected nonce %d, got %d", expectedNonce, account.Nonce)
		}

		if account.Balance.Cmp(expectedBalance) != 0 {
			t.Errorf("expected balance %s, got %s", expectedBalance.String(), account.Balance.String())
		}

		if account.CodeHash != expectedCodeHash {
			t.Errorf("expected code hash %x, got %x", expectedCodeHash, account.CodeHash)
		}

		if account.StorageRoot != expectedStorageRoot {
			t.Errorf("expected storage root %x, got %x", expectedStorageRoot, account.StorageRoot)
		}
	})

	t.Run("should return error on incomplete account proof", func(t *testing.T) {
		stateRoot := common.HexToHash("0x0136b96aa9d793cdccd5d1f4f03a576b0f64ce562dcb8d423414b5cff37e3d6c")
		address := common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266")

		_, err := VerifyAccountProof(stateRoot, address, make([][]byte, 0))
		if err == nil {
			t.Errorf("expected incomplete proof error")
		}
	})

	t.Run("should return error on corrupted account proof", func(t *testing.T) {
		stateRoot := common.HexToHash("0x0136b96aa9d793cdccd5d1f4f03a576b0f64ce562dcb8d423414b5cff37e3d6c")
		address := common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266")
		proof := []string{
			"0xf90131a0b91a8b7a7e9d3eab90afd81da3725030742f663c6ed8c26657bf00d842a9f4aaa01689b2a5203afd9ea0a0ca3765e4a538c7176e53eac1f8307a344ffc3c6176558080a0de673157fb5e8d14d783c948b64074922bf60224389cb46a3d38d48a7e81ae4ea04d5794121ef1a51608fa5b655bb3f861fb0a4fcecf8b7fecbf084b2d422a8bcf8080a04b29efa44ecf50c19b34950cf1d0f05e00568bcc873120fbea9a4e8439de0962a0d0a1bfe5b45d2d863a794f016450a4caca04f3b599e8d1652afca8b752935fd880a0bf9b09e442e044778b354abbadb5ec049d7f5e8b585c3966d476c4fbc9a181d28080a0a3a8f2834a8836fa2e4824f6c1dbe936a895fcfd53965acdf896567b138b90f6a0e5c557a0ce3894afeb44c37f3d24247f67dc76a174d8cacc360c1210eef60a7680",
			"0xf8518080808080a0aabfb1441169c3379f428df147ba34658049e31ab75bca31dcea5ea3513408a7808080a0df27128ae81e00b9ab17d7c0ff1fe52aa0320efba06361a8d6e9934daa27e76080808080808080",
			"0xf873a020707d0e6171f728f7473c24cc0432a9b07eaaf1efed6a137a4a8c12c79552d9b850f84e018a021e19e053fa587ede00a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a0c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
		}

		proofNodes := make([][]byte, len(proof))
		for idx, node := range proof {
			bytez, err := hex.DecodeString(strings.TrimPrefix(node, "0x"))
			if err != nil {
				t.Fatalf("failed to decode node %d %v", idx, node)
			}

			proofNodes[idx] = bytez
		}

		// Corrupt proof by changing the last byte of the last node
		proofNodes[len(proofNodes)-1][len(proofNodes[len(proofNodes)-1])-1] ^= 0x01

		_, err := VerifyAccountProof(stateRoot, address, proofNodes)
		if err == nil {
			t.Errorf("expected invalid proof error")
		}
	})
}

func TestVerifyStorageProof(t *testing.T) {
	t.Run("should verify valid one element storage proof", func(t *testing.T) {
		storageRoot := common.HexToHash("0xf258b1c6d5ee7f6f3549117fb0ac79118d5ad19b9c027f9b8e1471ca519b3b6c")
		paddedSlotZero := make([]byte, 32)
		slotKey := crypto.Keccak256Hash(paddedSlotZero)
		proof := []string{
			"0xe3a120290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e56307",
		}

		proofNodes := make([][]byte, len(proof))
		for idx, node := range proof {
			bytez, err := hex.DecodeString(strings.TrimPrefix(node, "0x"))
			if err != nil {
				t.Fatalf("failed to decode node %d %v", idx, node)
			}

			proofNodes[idx] = bytez
		}

		value, err := VerifyStorageProof(storageRoot, slotKey, proofNodes)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		expectedValue := []byte{0x07}
		if !bytes.Equal(expectedValue, value) {
			t.Errorf("expected storage value %x, got %x", expectedValue, value)
		}
	})

	t.Run("should verify valid storage proof for max uint256 element", func(t *testing.T) {
		storageRoot := common.HexToHash("0x5d85aa66d143fa6ff0a15deb90410bd8cd5a973c317d32d4d21e7731f73f35d2")
		slotTwo := big.NewInt(2)
		slotKey := crypto.Keccak256Hash(common.LeftPadBytes(slotTwo.Bytes(), 32))
		proof := []string{
			"0xf8718080a04355bd3061ad2d17e0782413925b4fd81a56bd162d91eedb2a00d6c87611471480a05cec288029f80518906c03ad962a0d47ecdf98680e3d85558885e7f3e7ac4bee808080808080a0df88c3b964bcf271c4442d25b05557a4baae5d23952f3c1d0149139f7127c68f8080808080",
			"0xf843a0305787fa12a823e0f2b7631cc41b3ba8828b3321ca811111fa75cd3aa3bb5acea1a0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		}

		proofNodes := make([][]byte, len(proof))
		for idx, node := range proof {
			bytez, err := hex.DecodeString(strings.TrimPrefix(node, "0x"))
			if err != nil {
				t.Fatalf("failed to decode node %d %v", idx, node)
			}

			proofNodes[idx] = bytez
		}

		value, err := VerifyStorageProof(storageRoot, slotKey, proofNodes)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		expectedValue := bytes.Repeat([]byte{0xff}, 32)
		if !bytes.Equal(expectedValue, value) {
			t.Errorf("expected storage value %x, got %x", expectedValue, value)
		}
	})

	t.Run("should return error on corrupted storage proof", func(t *testing.T) {
		storageRoot := common.HexToHash("0xf258b1c6d5ee7f6f3549117fb0ac79118d5ad19b9c027f9b8e1471ca519b3b6c")
		paddedSlotZero := make([]byte, 32)
		slotKey := crypto.Keccak256Hash(paddedSlotZero)
		proof := []string{
			"0xe3a120290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e56307",
		}

		proofNodes := make([][]byte, len(proof))
		for idx, node := range proof {
			bytez, err := hex.DecodeString(strings.TrimPrefix(node, "0x"))
			if err != nil {
				t.Fatalf("failed to decode node %d %v", idx, node)
			}

			proofNodes[idx] = bytez
		}

		// Corrupt proof by changing the last byte of the last node
		proofNodes[len(proofNodes)-1][len(proofNodes[len(proofNodes)-1])-1] ^= 0x01

		_, err := VerifyStorageProof(storageRoot, slotKey, proofNodes)
		if err == nil {
			t.Errorf("expected invalid proof error")
		}
	})
}

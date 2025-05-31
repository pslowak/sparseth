package ethclient

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"sparseth/execution/mpt"
)

// storageProvider provides verified Ethereum
// smart contract storage values.
type storageProvider struct {
	c *Client
}

// newStorageProvider creates a new storageProvider
// using the specified client.
func newStorageProvider(client *Client) *storageProvider {
	return &storageProvider{
		c: client,
	}
}

// getSlot provides the verified value stored at the
// specified storage slot for the specified Ethereum
// account at the specified block.
func (r *storageProvider) getSlot(ctx context.Context, account common.Address, slot common.Hash, header *types.Header) ([]byte, error) {
	proof, err := r.c.GetProof(ctx, account, []common.Hash{slot}, header.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get Proof: %w", err)
	}

	acc, err := mpt.VerifyAccountProof(header.Root, account, proof.AccountProof)
	if err != nil {
		return nil, fmt.Errorf("failed to verify account: %w", err)
	}

	if len(proof.StorageProof) == 0 {
		return nil, fmt.Errorf("missing storage proof for slot")
	}

	slotHash := crypto.Keccak256Hash(slot.Bytes())
	val, err := mpt.VerifyStorageProof(acc.StorageRoot, slotHash, proof.StorageProof[0].Proof)
	if err != nil {
		return nil, fmt.Errorf("failed to verify storage: %w", err)
	}

	return val, nil
}

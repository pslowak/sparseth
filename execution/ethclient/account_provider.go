package ethclient

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"sparseth/execution/mpt"
)

// accountProvider provides verified
// account-related data via the
// Ethereum RPC API.
type accountProvider struct {
	c *Client
}

// newAccountProvider creates a new accountProvider
// using the specified client.
func newAccountProvider(client *Client) *accountProvider {
	return &accountProvider{
		c: client,
	}
}

// getAccountAtBlock provides the verified
// account at the specified block, or nil
// if no such account exists.
func (p *accountProvider) getAccountAtBlock(ctx context.Context, account common.Address, header *types.Header) (*Account, error) {
	proof, err := p.c.GetProof(ctx, account, nil, header.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get proof: %w", err)
	}

	acc, err := mpt.VerifyAccountProof(header.Root, account, proof.AccountProof)
	if err != nil {
		return nil, fmt.Errorf("failed to verify account: %w", err)
	}
	if acc == nil {
		// Account does not exist
		return nil, nil
	}

	return &Account{
		Address:     account,
		Nonce:       acc.Nonce,
		Balance:     acc.Balance,
		CodeHash:    acc.CodeHash,
		StorageRoot: acc.StorageRoot,
	}, nil
}

// getSlotAtBlock provides the verified value stored
// at the specified storage slot for the specified
// Ethereum account at the specified block.
//
// Note that the specified account must exist at the
// specified block, otherwise an error will be returned.
func (p *accountProvider) getSlotAtBlock(ctx context.Context, addr common.Address, slot common.Hash, header *types.Header) ([]byte, error) {
	proof, err := p.c.GetProof(ctx, addr, []common.Hash{slot}, header.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get proof: %w", err)
	}

	acc, err := mpt.VerifyAccountProof(header.Root, addr, proof.AccountProof)
	if err != nil {
		return nil, fmt.Errorf("failed to verify account: %w", err)
	}
	if acc == nil {
		return nil, fmt.Errorf("account %s does not exist at block %d", addr.Hex(), header.Number.Uint64())
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

// getCodeAtBlock provides the verified code of the
// specified Ethereum account at the specified block.
//
// Note that the specified account must exist at the
// specified block, otherwise an error will be returned.
func (p *accountProvider) getCodeAtBlock(ctx context.Context, account common.Address, header *types.Header) ([]byte, error) {
	code, err := p.c.GetCodeAtBlock(ctx, account, header.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to get code at block: %w", err)
	}

	acc, err := p.getAccountAtBlock(ctx, account, header)
	if err != nil {
		return nil, fmt.Errorf("failed to get account at block: %w", err)
	}
	if acc == nil {
		return nil, fmt.Errorf("account %s does not exist at block %d", account.Hex(), header.Number.Uint64())
	}

	if acc.CodeHash != crypto.Keccak256Hash(code) {
		return nil, fmt.Errorf("account code hash does not match code")
	}

	return code, nil
}

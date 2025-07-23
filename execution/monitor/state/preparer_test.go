package state

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"log/slog"
	"math/big"
	"sparseth/execution/ethclient"
	"sparseth/internal/config"
	"sparseth/internal/log"
	"testing"
)

type preparerTestProvider struct {
	// access list to be returned by CreateAccessList
	al *types.AccessList
	// error to be returned by provider methods
	err error
}

func (p preparerTestProvider) GetTxsAtBlock(ctx context.Context, header *types.Header) ([]*ethclient.TransactionWithIndex, error) {
	return nil, nil
}

func (p preparerTestProvider) GetLogsAtBlock(ctx context.Context, acc common.Address, blockNum *big.Int) ([]*types.Log, error) {
	return nil, nil
}

func (p preparerTestProvider) GetAccountAtBlock(ctx context.Context, acc common.Address, head *types.Header) (*ethclient.Account, error) {
	return nil, nil
}

func (p preparerTestProvider) GetStorageAtBlock(ctx context.Context, acc common.Address, slot common.Hash, head *types.Header) ([]byte, error) {
	return nil, nil
}

func (p preparerTestProvider) GetCodeAtBlock(ctx context.Context, acc common.Address, head *types.Header) ([]byte, error) {
	return nil, nil
}

func (p preparerTestProvider) CreateAccessList(ctx context.Context, tx *ethclient.TransactionWithSender, blockNum *big.Int) (*types.AccessList, error) {
	return p.al, p.err
}

func TestPreparer_FilterTxs(t *testing.T) {
	testLogger := log.New(slog.DiscardHandler)

	t.Run("should return error when no access list could be retrieved", func(t *testing.T) {
		provider := preparerTestProvider{
			err: fmt.Errorf("failed to create access list"),
		}

		sk, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("failed to generate secret key: %v", err)
		}

		accs := &config.AccountsConfig{
			Accounts: []*config.AccountConfig{
				{
					Addr: crypto.PubkeyToAddress(sk.PublicKey),
				},
			},
		}

		header := &types.Header{Number: big.NewInt(1),
			Time: 1,
		}

		cc := params.TestChainConfig
		txData := &types.DynamicFeeTx{
			To:        &common.Address{},
			Value:     big.NewInt(1 * params.Ether),
			Nonce:     0,
			Gas:       21001,
			GasFeeCap: big.NewInt(1),
			GasTipCap: big.NewInt(2000000001),
		}
		signer := types.LatestSigner(cc)
		signedTx, err := types.SignNewTx(sk, signer, txData)
		if err != nil {
			t.Fatalf("failed to sign transaction: %v", err)
		}
		txs := []*ethclient.TransactionWithIndex{
			{
				Tx:    signedTx,
				Index: 0,
			},
		}

		preparer := NewPreparer(provider, nil, accs, cc, testLogger)
		filtered, err := preparer.FilterTxs(t.Context(), header, txs)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if filtered != nil {
			t.Errorf("expected no filtered transactions, got: %d", len(filtered))
		}
	})

	t.Run("should not filter tx when contract creation", func(t *testing.T) {
		provider := preparerTestProvider{
			al: &types.AccessList{},
		}

		sk, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("failed to generate secret key: %v", err)
		}

		accs := &config.AccountsConfig{
			Accounts: []*config.AccountConfig{},
		}

		header := &types.Header{Number: big.NewInt(1),
			Time: 1,
		}

		cc := params.TestChainConfig
		txData := &types.DynamicFeeTx{
			To:        nil, // contract creation
			Nonce:     1,
			Gas:       211052,
			GasFeeCap: big.NewInt(1875175000),
			Data:      []byte("6080604052348015600e575f5ffd5b506102db8061001c5f395ff3fe608060405234801561000f575f5ffd5b5060043610610034575f3560e01c80632e64cec1146100385780636057361d14610056575b5f5ffd5b610040610072565b60405161004d9190610133565b60405180910390f35b610070600480360381019061006b919061017a565b61007b565b005b5f600254905090565b60015f81548092919061008d906101d2565b91905055505f5433826040516020016100a893929190610270565b604051602081830303815290604052805190602001205f81905550806002819055503373ffffffffffffffffffffffffffffffffffffffff167f9372632017bf50244796e610d34ceaa5fb91a88d2b0bf3bb83cee5d957aa6e27826040516101109190610133565b60405180910390a250565b5f819050919050565b61012d8161011b565b82525050565b5f6020820190506101465f830184610124565b92915050565b5f5ffd5b6101598161011b565b8114610163575f5ffd5b50565b5f8135905061017481610150565b92915050565b5f6020828403121561018f5761018e61014c565b5b5f61019c84828501610166565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f6101dc8261011b565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff820361020e5761020d6101a5565b5b600182019050919050565b5f819050919050565b61022b81610219565b82525050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61025a82610231565b9050919050565b61026a81610250565b82525050565b5f6060820190506102835f830186610222565b6102906020830185610261565b61029d6040830184610124565b94935050505056fea26469706673582212204e429d361d5ba67ed310ddb7423554f0accb851bd6d05c454d6f0f8cf92312de64736f6c634300081e0033"),
		}
		signer := types.LatestSigner(cc)
		signedTx, err := types.SignNewTx(sk, signer, txData)
		if err != nil {
			t.Fatalf("failed to sign transaction: %v", err)
		}
		txs := []*ethclient.TransactionWithIndex{
			{
				Tx:    signedTx,
				Index: 0,
			},
		}

		preparer := NewPreparer(provider, nil, accs, cc, testLogger)
		filtered, err := preparer.FilterTxs(t.Context(), header, txs)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(filtered) != 1 {
			t.Errorf("exptected 1 filtered transaction, got: %d", len(filtered))
		}
	})

	t.Run("should not filter tx when sender is monitored", func(t *testing.T) {
		provider := preparerTestProvider{
			al: &types.AccessList{},
		}

		sk, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("failed to generate secret key: %v", err)
		}

		accs := &config.AccountsConfig{
			Accounts: []*config.AccountConfig{
				{
					Addr: crypto.PubkeyToAddress(sk.PublicKey),
				},
			},
		}

		header := &types.Header{Number: big.NewInt(1),
			Time: 1,
		}

		cc := params.TestChainConfig
		txData := &types.DynamicFeeTx{
			To:        &common.Address{},
			Value:     big.NewInt(1 * params.Ether),
			Nonce:     0,
			Gas:       21001,
			GasFeeCap: big.NewInt(1),
			GasTipCap: big.NewInt(2000000001),
		}
		signer := types.LatestSigner(cc)
		signedTx, err := types.SignNewTx(sk, signer, txData)
		if err != nil {
			t.Fatalf("failed to sign transaction: %v", err)
		}
		txs := []*ethclient.TransactionWithIndex{
			{
				Tx:    signedTx,
				Index: 0,
			},
		}

		preparer := NewPreparer(provider, nil, accs, cc, testLogger)
		filtered, err := preparer.FilterTxs(t.Context(), header, txs)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(filtered) != 1 {
			t.Errorf("exptected 1 filtered transaction, got: %d", len(filtered))
		}
	})

	t.Run("should not filter tx when receiver is monitored", func(t *testing.T) {
		provider := preparerTestProvider{
			al: &types.AccessList{},
		}

		sk, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("failed to generate secret key: %v", err)
		}

		rcvr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		accs := &config.AccountsConfig{
			Accounts: []*config.AccountConfig{
				{
					Addr: rcvr,
				},
			},
		}

		header := &types.Header{Number: big.NewInt(1),
			Time: 1,
		}

		cc := params.TestChainConfig
		txData := &types.DynamicFeeTx{
			To:        &rcvr,
			Value:     big.NewInt(1 * params.Ether),
			Nonce:     0,
			Gas:       21001,
			GasFeeCap: big.NewInt(1),
			GasTipCap: big.NewInt(2000000001),
		}
		signer := types.LatestSigner(cc)
		signedTx, err := types.SignNewTx(sk, signer, txData)
		if err != nil {
			t.Fatalf("failed to sign transaction: %v", err)
		}
		txs := []*ethclient.TransactionWithIndex{
			{
				Tx:    signedTx,
				Index: 0,
			},
		}

		preparer := NewPreparer(provider, nil, accs, cc, testLogger)
		filtered, err := preparer.FilterTxs(t.Context(), header, txs)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(filtered) != 1 {
			t.Errorf("exptected 1 filtered transaction, got: %d", len(filtered))
		}
	})

	t.Run("should not filter tx when monitored account is in access list", func(t *testing.T) {
		contract := common.HexToAddress("0x1234567890123456789012345678901234567890")
		provider := preparerTestProvider{
			al: &types.AccessList{
				{
					Address: contract,
					StorageKeys: []common.Hash{
						common.BigToHash(big.NewInt(0)),
					},
				},
			},
		}

		sk, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("failed to generate secret key: %v", err)
		}

		rcvr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		accs := &config.AccountsConfig{
			Accounts: []*config.AccountConfig{
				{
					Addr: contract,
				},
			},
		}

		header := &types.Header{Number: big.NewInt(1),
			Time: 1,
		}

		cc := params.TestChainConfig
		txData := &types.DynamicFeeTx{
			To:        &rcvr,
			Value:     big.NewInt(1 * params.Ether),
			Nonce:     0,
			Gas:       47963,
			GasTipCap: big.NewInt(1589011824),
			Data:      []byte("462b537c0000000000000000000000009fe46736679d2d9a65f0992f2272de9f3c7fa6e00000000000000000000000000000000000000000000000000000000000000001"),
		}
		signer := types.LatestSigner(cc)
		signedTx, err := types.SignNewTx(sk, signer, txData)
		if err != nil {
			t.Fatalf("failed to sign transaction: %v", err)
		}
		txs := []*ethclient.TransactionWithIndex{
			{
				Tx:    signedTx,
				Index: 0,
			},
		}

		preparer := NewPreparer(provider, nil, accs, cc, testLogger)
		filtered, err := preparer.FilterTxs(t.Context(), header, txs)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(filtered) != 1 {
			t.Errorf("exptected 1 filtered transaction, got: %d", len(filtered))
		}
	})

	t.Run("should not filter tx when receiver is monitored and sender sent tx earlier", func(t *testing.T) {
		provider := preparerTestProvider{
			al: &types.AccessList{},
		}

		sk, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("failed to generate secret key: %v", err)
		}

		rcvr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		accs := &config.AccountsConfig{
			Accounts: []*config.AccountConfig{
				{
					Addr: rcvr,
				},
			},
		}

		header := &types.Header{Number: big.NewInt(1),
			Time: 1,
		}

		cc := params.TestChainConfig
		signer := types.LatestSigner(cc)
		firstTxData := &types.DynamicFeeTx{
			To:        &common.Address{},
			Value:     big.NewInt(1 * params.Ether),
			Nonce:     0,
			Gas:       21001,
			GasFeeCap: big.NewInt(1),
			GasTipCap: big.NewInt(2000000001),
		}
		firstSignedTx, err := types.SignNewTx(sk, signer, firstTxData)
		secondTxData := &types.DynamicFeeTx{
			To:        &rcvr,
			Value:     big.NewInt(1 * params.Ether),
			Nonce:     0,
			Gas:       21001,
			GasFeeCap: big.NewInt(1),
			GasTipCap: big.NewInt(2000000001),
		}
		secondSignedTx, err := types.SignNewTx(sk, signer, secondTxData)
		if err != nil {
			t.Fatalf("failed to sign transaction: %v", err)
		}

		txs := []*ethclient.TransactionWithIndex{
			{
				Tx:    firstSignedTx,
				Index: 0,
			},
			{
				Tx:    secondSignedTx,
				Index: 1,
			},
		}

		preparer := NewPreparer(provider, nil, accs, cc, testLogger)
		filtered, err := preparer.FilterTxs(t.Context(), header, txs)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(filtered) != 2 {
			t.Errorf("exptected 2 filtered transactions, got: %d", len(filtered))
		}
	})

	t.Run("should filter tx when receiver is monitored and sender sent tx later", func(t *testing.T) {
		provider := preparerTestProvider{
			al: &types.AccessList{},
		}

		sk, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("failed to generate secret key: %v", err)
		}

		rcvr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		accs := &config.AccountsConfig{
			Accounts: []*config.AccountConfig{
				{
					Addr: rcvr,
				},
			},
		}

		header := &types.Header{Number: big.NewInt(1),
			Time: 1,
		}

		cc := params.TestChainConfig
		signer := types.LatestSigner(cc)
		firstTxData := &types.DynamicFeeTx{
			To:        &rcvr,
			Value:     big.NewInt(1 * params.Ether),
			Nonce:     0,
			Gas:       21001,
			GasFeeCap: big.NewInt(1),
			GasTipCap: big.NewInt(2000000001),
		}
		firstSignedTx, err := types.SignNewTx(sk, signer, firstTxData)
		secondTxData := &types.DynamicFeeTx{
			To:        &common.Address{},
			Value:     big.NewInt(1 * params.Ether),
			Nonce:     0,
			Gas:       21001,
			GasFeeCap: big.NewInt(1),
			GasTipCap: big.NewInt(2000000001),
		}
		secondSignedTx, err := types.SignNewTx(sk, signer, secondTxData)
		if err != nil {
			t.Fatalf("failed to sign transaction: %v", err)
		}

		txs := []*ethclient.TransactionWithIndex{
			{
				Tx:    firstSignedTx,
				Index: 0,
			},
			{
				Tx:    secondSignedTx,
				Index: 1,
			},
		}

		preparer := NewPreparer(provider, nil, accs, cc, testLogger)
		filtered, err := preparer.FilterTxs(t.Context(), header, txs)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(filtered) != 1 {
			t.Errorf("exptected 1 filtered transaction, got: %d", len(filtered))
		}
	})
}

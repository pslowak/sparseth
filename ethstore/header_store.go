package ethstore

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"sparseth/storage"
	"sync"
)

var (
	// ErrHeaderNotFound is returned when a
	// requested header is not found in the
	// store.
	ErrHeaderNotFound = errors.New("header not found")
)

// HeaderStore provides thread-safe access
// to Ethereum block headers.
//
// Two key mappings are maintained:
//   - Block number -> header hash
//   - Header hash -> header
type HeaderStore struct {
	db storage.KeyValStore
	mu sync.RWMutex
}

// NewHeaderStore creates a new HeaderStore
// using the specified key-val store.
func NewHeaderStore(db storage.KeyValStore) *HeaderStore {
	return &HeaderStore{
		db: db,
	}
}

// GetByHash retrieves a header by its hash.
func (s *HeaderStore) GetByHash(hash common.Hash) (*types.Header, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, err := s.db.Get(headerHashKey(hash))
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return nil, ErrHeaderNotFound
		}
		return nil, err
	}

	var header types.Header
	if err = rlp.DecodeBytes(val, &header); err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}
	return &header, nil
}

// GetByNumber retrieves a header by its block number.
func (s *HeaderStore) GetByNumber(num uint64) (*types.Header, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, err := s.db.Get(headerNumberKey(num))
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return nil, ErrHeaderNotFound
		}
		return nil, err
	}

	hash := common.BytesToHash(val)
	header, err := s.GetByHash(hash)
	if err != nil {
		// Since we already have the hash, a
		// non-existent header would indicate
		// a data inconsistency in the store.
		return nil, fmt.Errorf("failed to get header by hash: %w", err)
	}
	return header, nil
}

// Put stores the specified header in the store.
func (s *HeaderStore) Put(header *types.Header) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	encoded, err := rlp.EncodeToBytes(header)
	if err != nil {
		return err
	}

	batch := s.db.NewBatchWithSize(2)
	if err = batch.Put(headerHashKey(header.Hash()), encoded); err != nil {
		return fmt.Errorf("failed to put header in batch: %w", err)
	}
	if err = batch.Put(headerNumberKey(header.Number.Uint64()), header.Hash().Bytes()); err != nil {
		return fmt.Errorf("failed to put header in batch: %w", err)
	}
	return batch.Write()
}

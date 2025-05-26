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
	// ErrLogNotFound is returned when a
	// requested log is not found in the
	// store.
	ErrLogNotFound = errors.New("log not found")
	// logPrefix is used to prefix all log
	// entries in the key-val store.
	logPrefix = "log"
)

// EventStore provides thread-safe
// storage of Ethereum event logs.
type EventStore struct {
	db storage.KeyValStore
	mu sync.RWMutex
}

// NewEventStore creates a new EventStore
// using the specified key-val store.
func NewEventStore(db storage.KeyValStore) *EventStore {
	return &EventStore{
		db: db,
	}
}

// Close closes the underlying
// key-val store.
func (s *EventStore) Close() error {
	return s.db.Close()
}

// GetLog retrieves a log by transaction
// hash and log index.
func (s *EventStore) GetLog(txHash common.Hash, logIndex uint) (*types.Log, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := logKey(txHash, logIndex)
	encoded, err := s.db.Get([]byte(key))
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return nil, ErrLogNotFound
		}
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	var log types.Log
	if err = rlp.DecodeBytes(encoded, &log); err != nil {
		return nil, fmt.Errorf("failed to decode log: %w", err)
	}

	return &log, nil
}

// PutAll stores the specified logs
// into the EventStore.
func (s *EventStore) PutAll(logs []*types.Log) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	batch := s.db.NewBatchWithSize(len(logs))

	for _, log := range logs {
		encoded, err := rlp.EncodeToBytes(log)
		if err != nil {
			return fmt.Errorf("failed to encode log: %w", err)
		}

		if err = batch.Put([]byte(logKey(log.TxHash, log.Index)), encoded); err != nil {
			return fmt.Errorf("failed to put log in batch: %w", err)
		}
	}

	return batch.Write()
}

// logKey generates a unique key for a log.
func logKey(txHash common.Hash, logIndex uint) string {
	return fmt.Sprintf("%s:%s:%d", logPrefix, txHash.Hex(), logIndex)
}

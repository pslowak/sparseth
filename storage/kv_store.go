package storage

import (
	"errors"
	"io"
)

var (
	// ErrDbClosed is returned when the
	//storage is already closed
	ErrDbClosed = errors.New("storage closed")

	// ErrKeyNotFound is returned if the requested
	// key is not found in the storage
	ErrKeyNotFound = errors.New("key not found")
)

// KeyValReader defines read operations
// for a key-val store.
type KeyValReader interface {
	// Has checks if the specified key is
	// present in the key-val store.
	Has(key []byte) (bool, error)

	// Get retrieves the specified key if
	// it is present in the key-val store.
	Get(key []byte) ([]byte, error)
}

// KeyValWriter defines write operations
// of the key val store.
type KeyValWriter interface {
	// Put inserts the specified key-val
	// pair into the key-val store.
	Put(key, value []byte) error

	// Delete removes the specified key
	// from the key-val store.
	Delete(key []byte) error
}

// KeyValStater defines status operations
// of the key val store.
type KeyValStater interface {
	// Stat returns statistic data of
	// the database.
	Stat() (string, error)
}

// KeyValSyncer defines sync operations
// of the key val store.
type KeyValSyncer interface {
	// SyncKeyValue ensures that all pending
	// writes are flushed to disk.
	SyncKeyValue() error
}

// Compacter defines compaction operations
// of the key val store.
type Compacter interface {
	// Compact flattens the underlying key-val
	// store for the given key range.
	//
	// If start is nil, the compaction starts
	// before all keys in the data store. If
	// limit is nil, the compaction ends after
	// the last key in the data store.
	Compact(start []byte, limit []byte) error
}

type KeyValStore interface {
	KeyValReader
	KeyValWriter
	KeyValStater
	KeyValSyncer
	Batcher
	Iteratee
	Compacter
	io.Closer
}

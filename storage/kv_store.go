package storage

import "errors"

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

	// PutBatch inserts multiple key-val
	// pairs into the key-val store.
	PutBatch(pairs map[string][]byte) error

	// Delete removes the specified key
	// from the key-val store.
	Delete(key []byte) error
}

type KeyValStore interface {
	KeyValReader
	KeyValWriter

	// Close closes the underlying
	// key-val store.
	Close() error
}

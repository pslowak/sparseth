package storage

import (
	"errors"
	"github.com/ethereum/go-ethereum/ethdb"
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

// KeyValSyncer defines sync operations
// of the key val store.
type KeyValSyncer interface {
	// SyncKeyValue ensures that all pending
	// writes are flushed to disk.
	SyncKeyValue() error
}

type KeyValStore interface {
	ethdb.KeyValueReader
	ethdb.KeyValueWriter
	ethdb.KeyValueStater
	KeyValSyncer
	ethdb.KeyValueRangeDeleter
	ethdb.Batcher
	ethdb.Iteratee
	ethdb.Compacter
	io.Closer
}

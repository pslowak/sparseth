package badger

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"sparseth/storage"
)

// Database is a badger key-val store.
type Database struct {
	db *badger.DB
}

// New creates a new badger datastore
// instance at the specified path.
func New(path string) (*Database, error) {
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	return &Database{db: db}, nil
}

// Close closes the underlying datastore.
func (db *Database) Close() error {
	return db.db.Close()
}

// Has checks if the specified key exists
// in the datastore.
func (db *Database) Has(key []byte) (bool, error) {
	err := db.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		return err
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return false, nil
	}
	return err == nil, err
}

// Get retrieves the value associated with the
// specified key, if present.
func (db *Database) Get(key []byte) ([]byte, error) {
	var val []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, storage.ErrKeyNotFound
	}
	return val, err
}

// Put inserts the specified key-value pair
// into the datastore.
func (db *Database) Put(key, val []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

// Delete removes the specified key from
// the datastore.
func (db *Database) Delete(key []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Stat returns statistic data of
// the datastore.
func (db *Database) Stat() (string, error) {
	lsmSize, vlogSize := db.db.Size()
	return fmt.Sprintf("Badger DB lsm size: %d bytes, value log file size: %d bytes", lsmSize, vlogSize), nil
}

// SyncKeyValue ensures that all pending
// writes are flushed to disk.
func (db *Database) SyncKeyValue() error {
	return db.db.Sync()
}

// NewBatch creates a new write-only batch.
func (db *Database) NewBatch() storage.Batch {
	return &batch{
		db:  db,
		wb:  db.db.NewWriteBatch(),
		ops: make([]op, 0),
		sz:  0,
	}
}

// NewBatchWithSize creates a new batch with
// a pre-allocated buffer of the specified
// size.
func (db *Database) NewBatchWithSize(size int) storage.Batch {
	return &batch{
		db:  db,
		wb:  db.db.NewWriteBatch(),
		ops: make([]op, 0, size),
		sz:  0,
	}
}

// batch is a write-only batch for
// the badger datastore.
type batch struct {
	db  *Database
	wb  *badger.WriteBatch
	ops []op
	sz  int
}

type op struct {
	key []byte
	val []byte // nil if delete
	del bool
}

// Put inserts the specified key-value pair
// into the batch.
func (b *batch) Put(key, val []byte) error {
	if err := b.wb.Set(key, val); err != nil {
		return fmt.Errorf("failed to put key %s: %w", string(key), err)
	}
	b.ops = append(b.ops, op{
		key: copyBytes(key),
		val: copyBytes(val),
		del: false,
	})
	b.sz += len(key) + len(val)
	return nil
}

// Delete marks the specified key for deletion
// in the batch.
func (b *batch) Delete(key []byte) error {
	if err := b.wb.Delete(key); err != nil {
		return fmt.Errorf("failed to delete key %s: %w", string(key), err)
	}
	b.ops = append(b.ops, op{
		key: copyBytes(key),
		val: nil,
		del: true,
	})
	b.sz += len(key)
	return nil
}

// ValueSize retrieves the total size of data
// queued up for writing in the batch.
func (b *batch) ValueSize() int {
	return b.sz
}

// Write commits changes in the batch to the
// underlying datastore.
func (b *batch) Write() error {
	return b.wb.Flush()
}

// Reset clears the batch for reuse.
func (b *batch) Reset() {
	b.wb.Cancel()
	b.wb = b.db.db.NewWriteBatch()
	b.ops = b.ops[:0]
	b.sz = 0
}

// Replay replays the batch contents to the
// specified writer.
func (b *batch) Replay(w storage.KeyValWriter) error {
	for _, operation := range b.ops {
		if operation.del {
			if err := w.Delete(operation.key); err != nil {
				return fmt.Errorf("failed to delete key %s: %w", string(operation.key), err)
			}
		} else {
			if err := w.Put(operation.key, operation.val); err != nil {
				return fmt.Errorf("failed to put key %s: %w", string(operation.key), err)
			}
		}
	}

	return nil
}

// copyBytes creates a copy of the
// provided byte slice.
func copyBytes(b []byte) []byte {
	if b == nil {
		return nil
	}

	copied := make([]byte, len(b))
	copy(copied, b)
	return copied
}

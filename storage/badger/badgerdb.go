package badger

import (
	"bytes"
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

// DeleteRange deletes all keys (and values)
// in the range [start, end).
func (db *Database) DeleteRange(start, end []byte) error {
	err := db.db.Update(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := tx.NewIterator(opts)
		defer it.Close()

		for it.Seek(start); it.Valid(); it.Next() {
			key := it.Item().KeyCopy(nil)

			if bytes.Compare(key, end) >= 0 {
				break
			}

			if err := tx.Delete(key); err != nil {
				return fmt.Errorf("failed to delete key %s: %w", string(key), err)
			}
		}

		return nil
	})

	return err
}

// Compact flattens the database. In badger, value
// log file garbage collection is performed.
func (db *Database) Compact([]byte, []byte) error {
	if err := db.db.RunValueLogGC(0.5); err != nil {
		if errors.Is(err, badger.ErrNoRewrite) {
			// No compaction needed
			return nil
		}
		return fmt.Errorf("failed to compact value log: %w", err)
	}
	return nil
}

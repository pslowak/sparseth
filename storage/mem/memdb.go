package mem

import (
	"fmt"
	"sparseth/storage"
	"sync"
)

// Database is an in-memory key-value store.
type Database struct {
	db   map[string][]byte
	lock sync.RWMutex
}

// New creates a new in-memory database.
func New() *Database {
	return &Database{
		db: make(map[string][]byte),
	}
}

// Close deallocates the database. Any consecutive
// data access fails with an error.
func (db *Database) Close() error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.db = nil
	return nil
}

// Has checks if the specified key exists in
// the database.
func (db *Database) Has(key []byte) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if db.db == nil {
		return false, storage.ErrDbClosed
	}

	_, ok := db.db[string(key)]
	return ok, nil
}

// Get retrieves the value associated with the specified
// key, if present.
func (db *Database) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if db.db == nil {
		return nil, storage.ErrDbClosed
	}

	if val, ok := db.db[string(key)]; ok {
		return storage.CopyBytes(val), nil
	}

	return nil, storage.ErrKeyNotFound
}

// Put inserts the specified key-value pair into
// the database.
func (db *Database) Put(key, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.db == nil {
		return storage.ErrDbClosed
	}

	db.db[string(key)] = storage.CopyBytes(value)
	return nil
}

// Delete removes the specified key from the database.
func (db *Database) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.db == nil {
		return storage.ErrDbClosed
	}

	delete(db.db, string(key))
	return nil
}

// Stat returns statistic data of the database.
func (db *Database) Stat() (string, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if db.db == nil {
		return "", storage.ErrDbClosed
	}

	return fmt.Sprintf("Memory DB: %d keys stored", len(db.db)), nil
}

// SyncKeyValue SynKeyValue ensures that all
// pending writes are flushed to disk. In a
// memory database, this is a no-op.
func (db *Database) SyncKeyValue() error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.db == nil {
		return storage.ErrDbClosed
	}

	return nil
}

// DeleteRange deletes all keys (and values)
// in the range [start, end).
func (db *Database) DeleteRange(start, end []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.db == nil {
		return storage.ErrDbClosed
	}

	for key := range db.db {
		if key >= string(start) && key < string(end) {
			delete(db.db, key)
		}
	}

	return nil
}

// Compact flattens the database. In an
// in-memory database, this is a no-op.
func (db *Database) Compact([]byte, []byte) error {
	return nil
}

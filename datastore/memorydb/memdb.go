package memorydb

import (
	"sparseth/datastore"
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
		return false, datastore.ErrDbClosed
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
		return nil, datastore.ErrDbClosed
	}

	if val, ok := db.db[string(key)]; ok {
		return copyBytes(val), nil
	}

	return nil, datastore.ErrKeyNotFound
}

// Put inserts the specified key-value pair into
// the database.
func (db *Database) Put(key, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.db == nil {
		return datastore.ErrDbClosed
	}

	db.db[string(key)] = copyBytes(value)
	return nil
}

// Delete removes the specified key from the database.
func (db *Database) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.db == nil {
		return datastore.ErrDbClosed
	}

	delete(db.db, string(key))
	return nil
}

func copyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}

	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)
	return copiedBytes
}

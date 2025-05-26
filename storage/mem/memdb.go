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
		return copyBytes(val), nil
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

	db.db[string(key)] = copyBytes(value)
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

// NewBatch creates a new write-only batch.
func (db *Database) NewBatch() storage.Batch {
	return &batch{
		db:    db,
		pairs: make([]pair, 0),
		size:  0,
	}
}

// NewBatchWithSize creates a write-only batch
// with a pre-allocated buffer of the specified
// size.
func (db *Database) NewBatchWithSize(size int) storage.Batch {
	return &batch{
		db:    db,
		pairs: make([]pair, 0, size),
		size:  0,
	}
}

// pair is a single key-value pair.
type pair struct {
	key string
	val []byte
	del bool
}

// batch is a write-only collection of key-value
// pairs. Changes are reflected after the Write
// method is called. Note that batch is not safe
// for concurrent use.
type batch struct {
	db    *Database
	pairs []pair
	size  int
}

// Put inserts the specified key-value pair
// into the batch.
func (b *batch) Put(key, val []byte) error {
	item := pair{
		key: string(key),
		val: copyBytes(val),
		del: false,
	}

	b.pairs = append(b.pairs, item)
	b.size += len(key) + len(val)
	return nil
}

// Delete marks the specified key for deletion
// in the batch.
func (b *batch) Delete(key []byte) error {
	item := pair{
		key: string(key),
		val: nil,
		del: true,
	}

	b.pairs = append(b.pairs, item)
	b.size += len(key)
	return nil
}

// ValueSize retrieves the total size of data
// queued up for writing in the batch.
func (b *batch) ValueSize() int {
	return b.size
}

// Write commits changes in the batch to the
// underlying database.
func (b *batch) Write() error {
	b.db.lock.Lock()
	defer b.db.lock.Unlock()

	if b.db.db == nil {
		return storage.ErrDbClosed
	}

	for _, item := range b.pairs {
		if item.del {
			delete(b.db.db, item.key)
		} else {
			b.db.db[item.key] = item.val
		}
	}

	return nil
}

// Reset clears the batch for reuse.
func (b *batch) Reset() {
	b.pairs = b.pairs[:0]
	b.size = 0
}

// Replay replays the batch contents to
// the specified writer.
func (b *batch) Replay(w storage.KeyValWriter) error {
	for _, item := range b.pairs {
		if item.del {
			if err := w.Delete([]byte(item.key)); err != nil {
				return err
			}
		} else {
			if err := w.Put([]byte(item.key), item.val); err != nil {
				return err
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

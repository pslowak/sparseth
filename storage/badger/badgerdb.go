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

// NewIterator creates a binary-alphabetical
// iterator over a subset of the datastore's
// content with the specified key prefix,
// starting at the specified initial key.
func (db *Database) NewIterator(prefix, start []byte) storage.Iterator {
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix

	tx := db.db.NewTransaction(false)
	it := tx.NewIterator(opts)

	return &iterator{
		tx:     tx,
		it:     it,
		start:  append(prefix, start...),
		seeked: false,
	}
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

// iterator is a binary-alphabetical
// iterator over key-value pairs.
type iterator struct {
	tx     *badger.Txn
	it     *badger.Iterator
	start  []byte
	seeked bool
	err    error
}

// Next moves the iterator to the
// next key-value pair.
func (it *iterator) Next() bool {
	if !it.seeked {
		it.seeked = true
		it.it.Seek(it.start)
		return it.it.Valid()
	}

	if !it.it.Valid() {
		return false
	}

	it.it.Next()
	return it.it.Valid()
}

// Error returns any accumulated error
// during iteration.
func (it *iterator) Error() error {
	return it.err
}

// Key returns the key of the current
// key-value pair, or nil if the iterator
// is already exhausted.
func (it *iterator) Key() []byte {
	if !it.it.Valid() {
		return nil
	}
	return it.it.Item().KeyCopy(nil)
}

// Value returns the value of the current
// key-value pair, or nil if the iterator
// is already exhausted.
func (it *iterator) Value() []byte {
	if !it.it.Valid() {
		return nil
	}
	val, err := it.it.Item().ValueCopy(nil)
	if err != nil {
		it.err = fmt.Errorf("failed to get value: %w", err)
		return nil
	}
	return val
}

// Release releases associated resources.
func (it *iterator) Release() {
	it.it.Close()
	it.tx.Discard()

	// Hint GC
	it.it = nil
	it.tx = nil
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

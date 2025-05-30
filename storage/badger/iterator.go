package badger

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/ethdb"
)

// iterator is a binary-alphabetical
// iterator over key-value pairs.
type iterator struct {
	tx     *badger.Txn
	it     *badger.Iterator
	start  []byte
	seeked bool
	err    error
}

// NewIterator creates a binary-alphabetical
// iterator over a subset of the datastore's
// content with the specified key prefix,
// starting at the specified initial key.
func (db *Database) NewIterator(prefix, start []byte) ethdb.Iterator {
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

package badger

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/ethdb"
	"sparseth/storage"
)

// batch is a write-only batch for
// the badger datastore.
type batch struct {
	db  *Database
	wb  *badger.WriteBatch
	ops []*op
	sz  int
}

// op represents a single
// write operation.
type op struct {
	key []byte
	val []byte // nil if delete
	del bool
}

// NewBatch creates a new write-only batch.
func (db *Database) NewBatch() ethdb.Batch {
	return &batch{
		db:  db,
		wb:  db.db.NewWriteBatch(),
		ops: make([]*op, 0),
		sz:  0,
	}
}

// NewBatchWithSize creates a new batch with
// a pre-allocated buffer of the specified
// size.
func (db *Database) NewBatchWithSize(size int) ethdb.Batch {
	return &batch{
		db:  db,
		wb:  db.db.NewWriteBatch(),
		ops: make([]*op, 0, size),
		sz:  0,
	}
}

// Put inserts the specified key-value pair
// into the batch.
func (b *batch) Put(key, val []byte) error {
	if err := b.wb.Set(key, val); err != nil {
		return fmt.Errorf("failed to put key %s: %w", string(key), err)
	}
	b.ops = append(b.ops, &op{
		key: storage.CopyBytes(key),
		val: storage.CopyBytes(val),
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
	b.ops = append(b.ops, &op{
		key: storage.CopyBytes(key),
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
func (b *batch) Replay(w ethdb.KeyValueWriter) error {
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

package mem

import (
	"github.com/ethereum/go-ethereum/ethdb"
	"sparseth/storage"
)

// batch is a write-only collection of key-value
// pairs. Changes are reflected after the Write
// method is called. Note that batch is not safe
// for concurrent use.
type batch struct {
	db    *Database
	pairs []*pair
	size  int
}

// NewBatch creates a new write-only batch.
func (db *Database) NewBatch() ethdb.Batch {
	return &batch{
		db:    db,
		pairs: make([]*pair, 0),
		size:  0,
	}
}

// NewBatchWithSize creates a write-only batch
// with a pre-allocated buffer of the specified
// size.
func (db *Database) NewBatchWithSize(size int) ethdb.Batch {
	return &batch{
		db:    db,
		pairs: make([]*pair, 0, size),
		size:  0,
	}
}

// Put inserts the specified key-value pair
// into the batch.
func (b *batch) Put(key, val []byte) error {
	item := &pair{
		key: string(key),
		val: storage.CopyBytes(val),
		del: false,
	}

	b.pairs = append(b.pairs, item)
	b.size += len(key) + len(val)
	return nil
}

// Delete marks the specified key for deletion
// in the batch.
func (b *batch) Delete(key []byte) error {
	item := &pair{
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
func (b *batch) Replay(w ethdb.KeyValueWriter) error {
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

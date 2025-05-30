package mem

import (
	"github.com/ethereum/go-ethereum/ethdb"
	"sort"
	"sparseth/storage"
	"strings"
)

// iterator is a simple iterator over a (partial)
// keyspace of a memory key-value store.
type iterator struct {
	idx   int
	pairs []*pair
}

// NewIterator creates a binary-alphabetical
// iterator over a subset of the database
// content with the specified key prefix,
// starting at the specified initial key.
func (db *Database) NewIterator(prefix, start []byte) ethdb.Iterator {
	db.lock.RLock()
	defer db.lock.RUnlock()

	pr := string(prefix)
	st := string(append(prefix, start...))

	pairs := make([]*pair, 0, len(db.db))
	for k, v := range db.db {
		if strings.HasPrefix(k, pr) && k >= st {
			pairs = append(pairs, &pair{
				key: k,
				val: storage.CopyBytes(v),
			})
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].key < pairs[j].key
	})

	return &iterator{
		idx:   -1,
		pairs: pairs,
	}
}

// Next moves the iterator to the next
// key-value pair.
func (it *iterator) Next() bool {
	if it.idx >= len(it.pairs) {
		return false
	}

	it.idx++
	return it.idx < len(it.pairs)
}

// Error returns nil, as a memory iterator
// cannot encounter errors.
func (it *iterator) Error() error {
	return nil
}

// Key returns the key of the current
// key-value pair, or nil if the iterator
// is already exhausted.
func (it *iterator) Key() []byte {
	if it.idx < 0 || it.idx >= len(it.pairs) {
		return nil
	}

	return []byte(it.pairs[it.idx].key)
}

// Value returns the value of the current
// key-value pair, or nil if the iterator
// is already exhausted.
func (it *iterator) Value() []byte {
	if it.idx < 0 || it.idx >= len(it.pairs) {
		return nil
	}

	return it.pairs[it.idx].val
}

// Release releases associated resources.
func (it *iterator) Release() {
	it.idx = -1
	it.pairs = nil
}

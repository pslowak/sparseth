package storage

// Iterator iterates over the key-value pairs.
//
// An iterator must be released after use, but it is not
// necessary to read an iterator until exhaustion. Note
// that an iterator is not safe for concurrent use.
type Iterator interface {
	// Next moves the iterator to the next key-value
	// pair and reports weather the iterator is exhausted.
	Next() bool

	// Error returns any accumulated error.
	Error() error

	// Key returns the key of the current key-value
	// pair, or nil if no such key.
	Key() []byte

	// Value returns the value of the current key-value
	// pair, or nil if no such value.
	Value() []byte

	// Release releases associated resources.
	Release()
}

// Iteratee defines an Iteratee that can create
// an iterator.
type Iteratee interface {
	// NewIterator creates a binary-alphabetical
	// iterator over a subset with a specified
	// key prefix, starting at a specified
	// initial key (or after if it does not exist).
	// Note that the prefix is not part of the start.
	NewIterator(prefix []byte, start []byte) Iterator
}

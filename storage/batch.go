package storage

// Batch defines a write-only batch.
// Note that a batch is not safe for
// concurrent use.
type Batch interface {
	KeyValWriter

	// ValueSize returns the size of
	// the items queued for writing.
	ValueSize() int

	// Write commits changes in the batch
	// to the underlying datastore.
	Write() error

	// Reset clears the batch for reuse.
	Reset()

	// Replay replays the batch contents.
	Replay(w KeyValWriter) error
}

// Batcher defines a batcher that can
// create write-only batches.
type Batcher interface {
	// NewBatch creates a write-only batch that
	// buffers changes until write is called.
	NewBatch() Batch

	// NewBatchWithSize creates a write-only
	// database batch with the specified size.
	NewBatchWithSize(size int) Batch
}

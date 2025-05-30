package mem

// pair is a single key-value pair.
type pair struct {
	key string
	val []byte // nil if marked for deletion
	del bool
}

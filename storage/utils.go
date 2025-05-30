package storage

// CopyBytes creates a copy of the
// provided byte slice.
func CopyBytes(b []byte) []byte {
	if b == nil {
		return nil
	}

	copied := make([]byte, len(b))
	copy(copied, b)
	return copied
}

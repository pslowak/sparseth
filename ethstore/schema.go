package ethstore

import (
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common"
)

// Define low level database schema prefixes.
var (
	// sparsethPrefix is used to prefix all data stored
	// directly by the sparse node. This prefix is used
	// to avoid collisions with the go-ethereum lib.
	sparsethPrefix = []byte("se:")

	// logPrefix is used to prefix all log entries in
	// the key-val store.
	logPrefix = prefix("log:")
)

// logKey generates a unique key for a log.
//
// logKey = se:log:<txHash>:<logIndex>
func logKey(txHash common.Hash, logIndex uint) []byte {
	// 1 for the separator (':'), 8 for uint64
	key := make([]byte, 0, len(logPrefix)+common.HashLength+1+8)
	key = append(key, logPrefix...)
	key = append(key, txHash.Bytes()...)
	key = append(key, ':')
	key = append(key, encodeNumber(uint64(logIndex))...)
	return key
}

// prefix returns a byte slice that combines the
// sparsethPrefix with the specified string.
func prefix(s string) []byte {
	return append(sparsethPrefix, s...)
}

// encodeNumber encodes an uint64 number
// as big endian uint64.
func encodeNumber(num uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, num)
	return buf
}

package sketch

import (
	"encoding/binary"
	"hash/fnv"
)

func hash64(seed uint64, value string) uint64 {
	hasher := fnv.New64a()
	var seedBytes [8]byte
	binary.LittleEndian.PutUint64(seedBytes[:], seed)
	_, _ = hasher.Write(seedBytes[:])
	_, _ = hasher.Write([]byte(value))
	return hasher.Sum64()
}

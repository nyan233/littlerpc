package hash

import (
	"encoding/binary"
	"unsafe"
)

func Murmurhash3Onx8632(key []byte, seed uint32) uint32 {
	const (
		ChunkSize = 4
		C1        = 0xcc9e2d51
		C2        = 0x1b872593
		R1        = 15
		R2        = 13
		M         = 5
		N         = 0xe6546b64
	)
	hash := seed
	keyLen := len(key)
	nBlock := keyLen / ChunkSize
	for i := 0; i < nBlock; i++ {
		k := binary.LittleEndian.Uint32(key[i*ChunkSize : i*ChunkSize+ChunkSize])
		k = k * C1
		k = (k << R1) | (k >> (32 - R1))
		k = k * C2

		hash = hash ^ k
		hash = (hash << R2) | (hash >> (32 - R2))
		hash = hash*M + N
	}
	if nBlock*ChunkSize == keyLen {
		return hash
	}
	tail := (*[3]byte)(unsafe.Pointer(&key[nBlock*ChunkSize]))
	var k1 uint32
	switch keyLen & 3 {
	case 3:
		k1 ^= uint32(tail[2]) << 16
	case 2:
		k1 ^= uint32(tail[1]) << 8
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= C1
		k1 = (hash << R1) | (hash >> (32 - R1))
		k1 *= C2
		hash ^= k1
	}
	hash = hash ^ uint32(keyLen)
	hash = hash ^ (hash >> 16)
	hash = hash * 0x85ebca6b
	hash = hash ^ (hash >> 13)
	hash = hash * 0xc2b2ae35
	hash = hash ^ (hash >> 16)
	return hash
}

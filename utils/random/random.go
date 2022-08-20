package random

import (
	"github.com/nyan233/littlerpc/utils/hash"
	"strings"
	"unsafe"
)

func RandomStringOnAscii(maxBytes uint32) string {
	nByte := int(hash.FastRandN(maxBytes))
	var sb strings.Builder
	sb.Grow(nByte)
	for i := 0; i < nByte; i++ {
		sb.WriteByte(byte(hash.FastRandN(26)) + 65)
	}
	return sb.String()
}

func RandomBytesOnAscii(maxBytes uint32) []byte {
	str := RandomStringOnAscii(maxBytes)
	return *(*[]byte)(unsafe.Pointer(&str))
}

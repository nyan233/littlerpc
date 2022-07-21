package hash

import (
	_ "unsafe"
)

//go:linkname FastRandN runtime.fastrandn
func FastRandN(max uint32) uint32

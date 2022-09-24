package random

import (
	_ "unsafe"
)

//go:linkname FastRandN runtime.fastrandn
func FastRandN(max uint32) uint32

//go:linkname FastRand runtime.fastrand
func FastRand() uint32

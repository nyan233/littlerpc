//go:build !amd64

package asm

import "runtime"

func PAUSE() {
	runtime.Gosched()
}

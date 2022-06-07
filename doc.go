package littlerpc

import "unsafe"

// 一个声明的过程看起来应该是这样的
// func (r receiver) MethodName(arg1 Type, arg2 Type ...) (rep Type/*Type ..., err error/*coder.Error) {}

type eface struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

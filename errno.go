package littlerpc

import (
	"github.com/nyan233/littlerpc/coder"
)

var (
	ErrJsonUnMarshal      = coder.NewError("json unmarshal failed", "")
	ErrMethodNoRegister   = coder.NewError("method no register", "")
	ErrElemTypeNoRegister = coder.NewError("elem type no register : ", "")
	ErrServer             = coder.NewError("server error: ", "")
	ErrCallArgsType       = coder.NewError("call arguments type error : ", "")
	Nil                   = coder.NewError("the error is nil", "")
)

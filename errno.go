package littlerpc

import (
	"github.com/nyan233/littlerpc/protocol"
)

var (
	ErrJsonUnMarshal      = protocol.NewError("json unmarshal failed", "")
	ErrMethodNoRegister   = protocol.NewError("method no register", "")
	ErrElemTypeNoRegister = protocol.NewError("elem type no register : ", "")
	ErrMessageFormat      = protocol.NewError("message format invalid", "")
	ErrBodyRead           = protocol.NewError("read body failed : readN == ", "")
	ErrServer             = protocol.NewError("server error: ", "")
	ErrCallArgsType       = protocol.NewError("call arguments type error : ", "")
	Nil                   = protocol.NewError("the error is nil", "")
)

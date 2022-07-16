package common

import (
	"github.com/nyan233/littlerpc/protocol"
)

var (
	ErrMethodNoRegister   = protocol.NewError("method no register", "")
	ErrElemTypeNoRegister = protocol.NewError("elem type no register : ", "")
	ErrMessageFormat      = protocol.NewError("message format invalid", "")
	ErrBodyRead           = protocol.NewError("read body failed : readN == ", "")
	ErrServer             = protocol.NewError("server error: ", "")
	ErrCallArgsType       = protocol.NewError("call arguments type error : ", "")
	ErrCodecMarshalError  = protocol.NewError("codec.MarshalError return one error : ", "")
	ErrNoInstance         = protocol.NewError("instance not found : ", "")
	ErrNoMethod           = protocol.NewError("method not found : ", "")
	Nil                   = protocol.NewError("the error is nil", "")
)

package common

import (
	"github.com/nyan233/littlerpc/protocol/error"
)

var (
	Success               = error.LNewBaseError(error.Success, "OK")
	ErrMethodNoRegister   = error.LNewBaseError(error.MethodNoRegister, "method no register")
	ErrElemTypeNoRegister = error.LNewBaseError(error.InstanceNoRegister, "elem type no register")
	ErrMessageFormat      = error.LNewBaseError(error.MessageDecodingFailed, "message format invalid")
	ErrServer             = error.LNewBaseError(error.ServerError, "server error")
	ErrCallArgsType       = error.LNewBaseError(error.CallArgsTypeErr, "call arguments type error")
	ErrCodecMarshalError  = error.LNewBaseError(error.CodecMarshalErr, "codec.MarshalError return one error")
)

package common

import (
	"github.com/nyan233/littlerpc/protocol/error"
)

var (
	Success               = error.LNewStdError(error.Success, "OK")
	ErrMethodNoRegister   = error.LNewStdError(error.MethodNoRegister, "method no register")
	ErrElemTypeNoRegister = error.LNewStdError(error.InstanceNoRegister, "elem type no register")
	ErrMessageFormat      = error.LNewStdError(error.MessageDecodingFailed, "message format invalid")
	ErrServer             = error.LNewStdError(error.ServerError, "server error")
	ErrCallArgsType       = error.LNewStdError(error.CallArgsTypeErr, "call arguments type error")
	ErrCodecMarshalError  = error.LNewStdError(error.CodecMarshalErr, "codec.MarshalError return one error")
)

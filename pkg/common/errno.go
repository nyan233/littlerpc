package common

import (
	"github.com/nyan233/littlerpc/protocol/error"
)

var (
	Success                = error.LNewStdError(error.Success, "OK")
	ErrMethodNoRegister    = error.LNewStdError(error.MethodNoRegister, "method no register")
	ErrElemTypeNoRegister  = error.LNewStdError(error.InstanceNoRegister, "elem type no register")
	ErrMessageDecoding     = error.LNewStdError(error.MessageDecodingFailed, "message decoding invalid")
	ErrMessageEncoding     = error.LNewStdError(error.MessageEncodingFailed, "message encoding invalid")
	ErrServer              = error.LNewStdError(error.ServerError, "server error")
	ErrCallArgsType        = error.LNewStdError(error.CallArgsTypeErr, "call arguments type error")
	ErrCodecMarshalError   = error.LNewStdError(error.CodecMarshalErr, "codec.Marshal return one error")
	ErrCodecUnMarshalError = error.LNewStdError(error.CodecMarshalErr, "codec.UnMarshal return one error")
	ErrClient              = error.LNewStdError(error.ClientError, "client error")
)

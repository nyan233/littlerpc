package errorhandler

import (
	"github.com/nyan233/littlerpc/protocol/error"
)

var (
	Success                = DefaultErrHandler.LNewErrorDesc(error.Success, "OK")
	ServiceNotfound        = DefaultErrHandler.LNewErrorDesc(error.ServiceNotFound, "service no register")
	ErrMessageDecoding     = DefaultErrHandler.LNewErrorDesc(error.MessageDecodingFailed, "RpcMessage decoding invalid")
	ErrMessageEncoding     = DefaultErrHandler.LNewErrorDesc(error.MessageEncodingFailed, "RpcMessage encoding invalid")
	ErrServer              = DefaultErrHandler.LNewErrorDesc(error.ServerError, "server error")
	ErrCallArgsType        = DefaultErrHandler.LNewErrorDesc(error.CallArgsTypeErr, "call arguments type error")
	ErrCodecMarshalError   = DefaultErrHandler.LNewErrorDesc(error.CodecMarshalErr, "codec.Marshal return one error")
	ErrCodecUnMarshalError = DefaultErrHandler.LNewErrorDesc(error.CodecMarshalErr, "codec.UnMarshal return one error")
	ErrClient              = DefaultErrHandler.LNewErrorDesc(error.ClientError, "client error")
	ErrConnection          = DefaultErrHandler.LNewErrorDesc(error.ConnectionErr, "connection error")
	ContextNotFound        = DefaultErrHandler.LNewErrorDesc(error.ContextNotFound, "context not found")
)

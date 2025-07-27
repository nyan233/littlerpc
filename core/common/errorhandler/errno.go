package errorhandler

import (
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
)

var (
	Success                = DefaultErrHandler.LNewErrorDesc(error2.Success, "OK")
	ServiceNotfound        = DefaultErrHandler.LNewErrorDesc(error2.ServiceNotFound, "service no register")
	ErrMessageDecoding     = DefaultErrHandler.LNewErrorDesc(error2.MessageDecodingFailed, "RpcMessage decoding invalid")
	ErrMessageEncoding     = DefaultErrHandler.LNewErrorDesc(error2.MessageEncodingFailed, "RpcMessage encoding invalid")
	ErrServer              = DefaultErrHandler.LNewErrorDesc(error2.ServerError, "server error")
	ErrCallArgsType        = DefaultErrHandler.LNewErrorDesc(error2.CallArgsTypeErr, "call arguments type error")
	ErrCodecMarshalError   = DefaultErrHandler.LNewErrorDesc(error2.CodecMarshalErr, "codec.Marshal return one error")
	ErrCodecUnMarshalError = DefaultErrHandler.LNewErrorDesc(error2.CodecMarshalErr, "codec.UnMarshal return one error")
	ErrClient              = DefaultErrHandler.LNewErrorDesc(error2.ClientError, "client error")
	ErrConnection          = DefaultErrHandler.LNewErrorDesc(error2.ConnectionErr, "connection error")
	ContextNotFound        = DefaultErrHandler.LNewErrorDesc(error2.ContextNotFound, "context not found")
	ErrPlugin              = DefaultErrHandler.LNewErrorDesc(10275, "plugin error")
	StreamIdNotfound       = DefaultErrHandler.LNewErrorDesc(10281, "stream not found")
)

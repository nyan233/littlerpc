package msgwriter

import (
	"encoding/base64"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/jsonrpc2"
	"github.com/nyan233/littlerpc/core/middle/codec"
	errno "github.com/nyan233/littlerpc/core/protocol/error"
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"strconv"
)

type JsonRPC2 struct {
	Codec codec.Codec
}

func NewJsonRPC2(writers ...Writer) Writer {
	return &JsonRPC2{Codec: codec.Get("json")}
}

func (j *JsonRPC2) Header() []byte {
	return []byte{jsonrpc2.Header}
}

func (j *JsonRPC2) Write(arg Argument, header byte) error2.LErrorDesc {
	switch arg.Message.GetMsgType() {
	case message.Call, message.ContextCancel:
		return j.requestWrite(arg)
	case message.Return:
		return j.responseWrite(arg)
	default:
		return arg.EHandle.LNewErrorDesc(error2.UnsafeOption,
			"jsonrpc2 not supported message type", arg.Message.GetMsgType())
	}
}

func (j *JsonRPC2) Reset() {
	return
}

func (j *JsonRPC2) requestWrite(arg Argument) error2.LErrorDesc {
	request := jsonrpc2.Request{
		BaseMessage: jsonrpc2.BaseMessage{
			MetaData: make(map[string]string, 16),
		},
	}
	request.MessageType = int(arg.Message.GetMsgType())
	request.Id = arg.Message.GetMsgId()
	request.Version = jsonrpc2.Version
	arg.Message.MetaData.Range(func(key, val string) bool {
		request.MetaData[key] = val
		return true
	})
	request.Method = arg.Message.GetServiceName()
	if arg.Encoder != nil && arg.Encoder.Scheme() != message.DefaultPacker {
		return arg.EHandle.LNewErrorDesc(errno.UnsafeOption, "usage not supported for packer, only support text")
	}
	// json是文本型数据, 不需要base64编码, 如果是protobuf等二进制则需要base64编码之后才能被解析
	var isJsonParams bool
	if scheme := arg.Message.MetaData.Load(message.CodecScheme); scheme == "" || scheme == message.DefaultCodec {
		isJsonParams = true
	}
	iter := arg.Message.PayloadsIterator()
	request.Params = append(request.Params, '[')
	for iter.Next() {
		if !isJsonParams {
			request.Params = append(request.Params, '"')
			request.Params = append(request.Params, base64.StdEncoding.EncodeToString(iter.Take())...)
			request.Params = append(request.Params, '"')
		} else {
			request.Params = append(request.Params, iter.Take()...)
		}
		if iter.Next() {
			request.Params = append(request.Params, ',')
		}
	}
	request.Params = append(request.Params, ']')
	bytes, err := j.Codec.Marshal(&request)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(errorhandler.ErrMessageEncoding,
			fmt.Sprintf("jsonrpc2 Marshal LRPCMessage failed: %v", err))
	}
	writeN, err := arg.Conn.Write(bytes)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(errorhandler.ErrConnection, fmt.Sprintf("jsonrpc2 write failed: %v", err))
	}
	if writeN != len(bytes) {
		return arg.EHandle.LWarpErrorDesc(errorhandler.ErrConnection, fmt.Sprintf("jsonrpc2 write no complete: (%d:%d)", writeN, len(bytes)))
	}
	return nil
}

func (j *JsonRPC2) responseWrite(arg Argument) error2.LErrorDesc {
	var rep jsonrpc2.Response
	rep.MessageType = int(arg.Message.GetMsgType())
	errCode := arg.Message.MetaData.Load(message.ErrorCode)
	errMessage := arg.Message.MetaData.Load(message.ErrorMessage)
	errMore := arg.Message.MetaData.Load(message.ErrorMore)
	if errCode != "" && errMessage != "" {
		rep.Error = &jsonrpc2.Error{}
		switch code, _ := strconv.Atoi(errCode); code {
		case error2.ServiceNotFound:
			rep.Error.Code = jsonrpc2.MethodNotFound
		case error2.Success:
			rep.Error = nil
			goto handleResult
		case error2.MessageDecodingFailed:
			rep.Error.Code = jsonrpc2.ErrorParser
		case error2.UnsafeOption:
			rep.Error.Code = jsonrpc2.ErrorInternal
		case error2.CallArgsTypeErr, error2.CodecMarshalErr:
			rep.Error.Code = jsonrpc2.InvalidParams
		case error2.Unknown:
			rep.Error.Code = jsonrpc2.Unknown
		}
		rep.Error.Message = errMessage
		rep.Error.Data = convert.StringToBytes(errMore)
	}
handleResult:
	iter := arg.Message.PayloadsIterator()
	for iter.Next() {
		rep.Result = append(rep.Result, iter.Take())
	}
	rep.Id = arg.Message.GetMsgId()
	rep.Version = jsonrpc2.Version
	bytes, err := j.Codec.Marshal(rep)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(errorhandler.ErrMessageEncoding,
			fmt.Sprintf("jsonrpc2 Marshal LRPCMessage failed: %v", err))
	}
	writeN, err := arg.Conn.Write(bytes)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(errorhandler.ErrConnection,
			fmt.Sprintf("JsonRpc2 NoMux write error: %v", err))
	}
	if writeN != len(bytes) {
		return arg.EHandle.LWarpErrorDesc(errorhandler.ErrConnection,
			fmt.Sprintf("JsonRpc2 NoMux write bytes not equal : w(%d) != b(%d)", writeN, len(bytes)))
	}
	return nil
}

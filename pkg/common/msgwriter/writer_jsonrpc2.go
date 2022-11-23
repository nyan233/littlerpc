package msgwriter

import (
	"encoding/base64"
	"fmt"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/jsonrpc2"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
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

func (j *JsonRPC2) Write(arg Argument, header byte) perror.LErrorDesc {
	switch arg.Message.GetMsgType() {
	case message.Call, message.ContextCancel:
		return j.requestWrite(arg)
	case message.Return:
		return j.responseWrite(arg)
	default:
		return arg.EHandle.LNewErrorDesc(perror.UnsafeOption,
			"jsonrpc2 not supported message type", arg.Message.GetMsgType())
	}
}

func (j *JsonRPC2) Reset() {
	return
}

func (j *JsonRPC2) requestWrite(arg Argument) perror.LErrorDesc {
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
	iter := arg.Message.PayloadsIterator()
	request.Params = append(request.Params, '[')
	for iter.Next() {
		var bytes []byte
		var err error
		if arg.Encoder.Scheme() != message.DefaultPacker {
			bytes, err = arg.Encoder.EnPacket(iter.Take())
			if err != nil {
				return arg.EHandle.LWarpErrorDesc(common.ErrMessageEncoding,
					fmt.Sprintf("jsonrpc2 Packer UnPacket failed: %v", err))
			}
		} else {
			bytes = iter.Take()
		}
		request.Params = append(request.Params, '"')
		request.Params = append(request.Params, base64.StdEncoding.EncodeToString(bytes)...)
		request.Params = append(request.Params, '"')
		if iter.Next() {
			request.Params = append(request.Params, ',')
		}
	}
	request.Params = append(request.Params, ']')
	bytes, err := j.Codec.Marshal(&request)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(common.ErrMessageEncoding,
			fmt.Sprintf("jsonrpc2 Marshal LRPCMessage failed: %v", err))
	}
	writeN, err := arg.Conn.Write(bytes)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(common.ErrConnection, fmt.Sprintf("jsonrpc2 write failed: %v", err))
	}
	if writeN != len(bytes) {
		return arg.EHandle.LWarpErrorDesc(common.ErrConnection, fmt.Sprintf("jsonrpc2 write no complete: (%d:%d)", writeN, len(bytes)))
	}
	return nil
}

func (j *JsonRPC2) responseWrite(arg Argument) perror.LErrorDesc {
	var rep jsonrpc2.Response
	rep.MessageType = int(arg.Message.GetMsgType())
	errCode := arg.Message.MetaData.Load(message.ErrorCode)
	errMessage := arg.Message.MetaData.Load(message.ErrorMessage)
	errMore := arg.Message.MetaData.Load(message.ErrorMore)
	if errCode != "" && errMessage != "" {
		rep.Error = &jsonrpc2.Error{}
		switch code, _ := strconv.Atoi(errCode); code {
		case perror.ServiceNotFound:
			rep.Error.Code = jsonrpc2.MethodNotFound
		case perror.Success:
			rep.Error = nil
			goto handleResult
		case perror.MessageDecodingFailed:
			rep.Error.Code = jsonrpc2.ErrorParser
		case perror.UnsafeOption:
			rep.Error.Code = jsonrpc2.ErrorInternal
		case perror.CallArgsTypeErr, perror.CodecMarshalErr:
			rep.Error.Code = jsonrpc2.InvalidParams
		case perror.Unknown:
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
		return arg.EHandle.LWarpErrorDesc(common.ErrMessageEncoding,
			fmt.Sprintf("jsonrpc2 Marshal LRPCMessage failed: %v", err))
	}
	writeN, err := arg.Conn.Write(bytes)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(common.ErrConnection,
			fmt.Sprintf("JsonRpc2 NoMux write error: %v", err))
	}
	if writeN != len(bytes) {
		return arg.EHandle.LWarpErrorDesc(common.ErrConnection,
			fmt.Sprintf("JsonRpc2 NoMux write bytes not equal : w(%d) != b(%d)", writeN, len(bytes)))
	}
	return nil
}

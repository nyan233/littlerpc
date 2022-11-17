package msgwriter

import (
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
	case message.Call:
		return j.requestWrite(arg)
	case message.Return:
		return j.responseWrite(arg)
	case message.ContextCancel:
		return nil
	default:
		return arg.EHandle.LNewErrorDesc(perror.UnsafeOption,
			"jsonrpc2 not supported message type", arg.Message.GetMsgType())
	}
}

func (j *JsonRPC2) Reset() {
	return
}

func (j *JsonRPC2) requestWrite(arg Argument) perror.LErrorDesc {
	return arg.EHandle.LNewErrorDesc(perror.UnsafeOption, "jsonrpc2 not supported request message type")
}

func (j *JsonRPC2) responseWrite(arg Argument) perror.LErrorDesc {
	var rep jsonrpc2.Response
	rep.MessageType = jsonrpc2.ResponseType
	errCode := arg.Message.MetaData.Load(message.ErrorCode)
	errMessage := arg.Message.MetaData.Load(message.ErrorMessage)
	errMore := arg.Message.MetaData.Load(message.ErrorMore)
	if errCode != "" && errMessage != "" {
		rep.Error = &jsonrpc2.Error{}
		switch code, _ := strconv.Atoi(errCode); code {
		case perror.ServiceNotFound:
			rep.Error.Code = jsonrpc2.MethodNotFound
		case perror.Success:
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
	rep.Id = int64(arg.Message.GetMsgId())
	rep.Version = jsonrpc2.Version
	bytes, err := j.Codec.Marshal(rep)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(common.ErrCodecMarshalError, j.Codec.Scheme(), err.Error())
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

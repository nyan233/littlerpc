package msgwriter

import (
	"encoding/json"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"net/http"
	"strconv"
)

const (
	ErrorParser    = -32700 // jsonrpc2 解析消息失败
	InvalidRequest = -32600 // 无效的请求
	MethodNotFound = -32601 // 找不到方法
	InvalidParams  = -32602 // 无效的参数
	ErrorInternal  = -32603 // 内部错误
	Unknown        = -32004 // 未知的错误
)

type JsonRPC2Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type JsonRPC2Response struct {
	Version string         `json:"jsonrpc"`
	Result  [][]byte       `json:"result"`
	Error   *JsonRPC2Error `json:"error,omitempty"`
	Id      int64          `json:"id"`
}

type JsonRPC2 struct {
	Codec codec.Codec
}

func (j *JsonRPC2) Writer(arg Argument) perror.LErrorDesc {
	repWriter, ok := arg.Conn.(http.ResponseWriter)
	if ok {
		repWriter.WriteHeader(http.StatusOK)
	}
	var rep JsonRPC2Response
	errCode := arg.Message.MetaData.Load(message.ErrorCode)
	errMessage := arg.Message.MetaData.Load(message.ErrorMessage)
	errMore := arg.Message.MetaData.Load(message.ErrorMore)
	if errCode != "" && errMessage != "" {
		rep.Error = &JsonRPC2Error{}
		switch code, _ := strconv.Atoi(errCode); code {
		case perror.MethodNoRegister, perror.InstanceNoRegister:
			rep.Error.Code = MethodNotFound
		case perror.Success:
			goto handleResult
		case perror.MessageDecodingFailed:
			rep.Error.Code = ErrorParser
		case perror.UnsafeOption:
			rep.Error.Code = ErrorInternal
		case perror.CallArgsTypeErr, perror.CodecMarshalErr:
			rep.Error.Code = InvalidParams
		case perror.Unknown:
			rep.Error.Code = Unknown
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
	rep.Version = msgparser.JsonRPC2Version
	bytes, err := j.Codec.Marshal(rep)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(common.ErrCodecMarshalError, j.Codec.Scheme(), err.Error())
	}
	err = common.WriteControl(arg.Conn, bytes)
	if err != nil {
		return arg.EHandle.LWarpErrorDesc(common.ErrConnection, err)
	}
	return nil
}

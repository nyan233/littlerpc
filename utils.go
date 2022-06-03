package littlerpc

import (
	"encoding/json"
	"errors"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/nyan233/littlerpc/coder"
	"runtime"
	"strconv"
)


func HandleError(sp coder.RStackFrame,errNo coder.Error,conn *websocket.Conn,appendInfo string, more ...interface{}) {
	md := coder.CalleeMd{
		ArgType:    coder.Struct,
		Rep:        nil,
	}
	switch errNo.Info {
	case ErrJsonUnMarshal.Info:
		_, file, line, _ := runtime.Caller(2)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Response = append(sp.Response,md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.WriteMessage(websocket.TextMessage,errNoBytes)
		break
	case ErrMethodNoRegister.Info:
		_, file, line, _ := runtime.Caller(2)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Response = append(sp.Response,md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.WriteMessage(websocket.TextMessage,errNoBytes)
		break
	case ErrServer.Info:
		errNo.Info += appendInfo
		_, file, line, _ := runtime.Caller(1)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Response = append(sp.Response,md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.WriteMessage(websocket.TextMessage,errNoBytes)
	case Nil.Info:
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Response = append(sp.Response,md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.WriteMessage(websocket.TextMessage,errNoBytes)
	}
}

func encodeAnyBytes() {}
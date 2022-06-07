package littlerpc

import (
	"encoding/json"
	"errors"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/nyan233/littlerpc/coder"
	"reflect"
	"runtime"
	"strconv"
)

func HandleError(sp coder.RStackFrame, errNo coder.Error, conn *websocket.Conn, appendInfo string, more ...interface{}) {
	md := coder.CalleeMd{
		ArgType:    coder.Struct,
		AppendType: coder.ServerError,
		Rep:        nil,
	}
	switch errNo.Info {
	case ErrJsonUnMarshal.Info, ErrMethodNoRegister.Info, ErrCallArgsType.Info:
		errNo.Trace += appendInfo
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Response = append(sp.Response, md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.WriteMessage(websocket.TextMessage, errNoBytes)
		break
	case ErrServer.Info:
		errNo.Info += appendInfo
		_, file, line, _ := runtime.Caller(1)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Response = append(sp.Response, md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.WriteMessage(websocket.TextMessage, errNoBytes)
	case Nil.Info:
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Response = append(sp.Response, md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.WriteMessage(websocket.TextMessage, errNoBytes)
	}
}

// Little-RPC中定义了传入类型中不能为指针类型，所以Server根据这种方法就能快速判断
// 序列化好的远程栈帧的每个帧的类型是否和需要调用的参数列表的每个参数的类型相同
// 如果inputS有receiver的话，需要调用者对slice做Offset，比如[1:]
func checkInputTypeList(callArgs []reflect.Value, inputS []interface{}) (bool, []string) {
	if len(callArgs) != len(inputS) {
		return false, nil
	}
	for k := range callArgs {
		if !(callArgs[k].Type().Kind() == reflect.TypeOf(inputS[k]).Kind()) {
			return false, []string{callArgs[k].Type().Kind().String(),
				reflect.TypeOf(inputS[k]).Kind().String()}
		}
	}
	return true, nil
}

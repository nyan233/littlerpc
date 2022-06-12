package littlerpc

import (
	"encoding/json"
	"fmt"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/nyan233/littlerpc/internal/transport"
	"github.com/nyan233/littlerpc/protocol"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"reflect"
)

// try 指示是否需要重入处理结果的逻辑
// cr2 表示内部append过的callResult，以使更改调用者可见
func (s *Server) handleErrAndRepResult(c *websocket.Conn,callResult []reflect.Value,rep *protocol.Message) (cr2 []reflect.Value,try bool){
	errMd := protocol.FrameMd{
		ArgType: protocol.Struct,
	}

	switch i := lreflect.ToValueTypeEface(callResult[len(callResult)-1]); i.(type) {
	case *protocol.Error:
		errBytes, err := json.Marshal(i)
		if err != nil {
			HandleError(*rep, *ErrServer, c, err.Error())
			return
		}
		errMd.ArgType = protocol.Struct
		errMd.Data = errBytes
	case error:
		any := protocol.AnyArgs{
			Any: i.(error).Error(),
		}
		anyBytes, err := json.Marshal(&any)
		if err != nil {
			return
		}
		errMd.ArgType = protocol.String
		errMd.Data = anyBytes
	case nil:
		any := protocol.AnyArgs{
			Any: 0,
		}
		errMd.ArgType = protocol.Integer
		anyBytes, err := json.Marshal(&any)
		if err != nil {
			HandleError(*rep, *ErrServer, c, err.Error())
			return
		}
		errMd.Data = anyBytes
	default:
		// 现在允许最后一个返回值不是*code.Error/error，这种情况被视为没有错误
		callResult = append(callResult, reflect.ValueOf(nil))
		// 返回值长度为一，且不是错误类型
		// 证明前面的结果处理可能没有处理这个结果，这时候往末尾添加一个无意义的值，让结果得到正确的处理
		if len(callResult) == 2 {
			return callResult,true
		}
		callResult = callResult[len(callResult)-2:]
		// 如果最后没有返回*code.Error/error会导致遗漏处理一些返回值
		// 这个时候需要重新检查
		return callResult,true
	}
	// TODO : implement text encoding and gzip encoding
	rep.Header.Encoding = protocol.DefaultEncodingType
	rep.Body.Frame = append(rep.Body.Frame, errMd)
	// write header
	buf := s.bufferPool.Get().(*transport.BufferPool)
	defer s.bufferPool.Put(buf)
	buf.Buf = buf.Buf[:0]
	buf.Buf = append(buf.Buf,writeHeader(rep.Header)...)
	// write body
	repBytes, err := json.Marshal(rep.Body)
	if err != nil {
		HandleError(*rep, *ErrServer, c, err.Error())
		return
	}
	bytes, err := s.encoder.EnPacket(repBytes)
	if err != nil {
		HandleError(*rep, *ErrServer, c, err.Error())
		return
	}
	buf.Buf = append(buf.Buf,bytes...)
	// write data
	err = c.WriteMessage(websocket.TextMessage, buf.Buf)
	if err != nil {
		s.logger.ErrorFromErr(err)
		return nil,false
	}
	return callResult,false
}

// 将用户过程的返回结果集序列化为可传输的json数据
func (s *Server) handleResult(c *websocket.Conn,callResult []reflect.Value,rep *protocol.Message) {
	for _, v := range callResult[:len(callResult)-1] {
		var md protocol.FrameMd
		var eface = v.Interface()
		typ := checkIType(eface)
		// 返回值的类型为指针的情况，为其设置参数类型和正确的附加类型
		if typ == protocol.Pointer {
			md.ArgType = checkIType(v.Elem().Interface())
			if md.ArgType == protocol.Map || md.ArgType == protocol.Struct {
				_ = true
			}
		} else {
			md.ArgType = typ
		}
		// Map/Struct也需要Any包装器
		any := protocol.AnyArgs{
			Any: eface,
		}
		anyBytes, err := json.Marshal(&any)
		if err != nil {
			HandleError(*rep, *ErrServer, c, "")
			return
		}
		md.Data = anyBytes
		rep.Body.Frame = append(rep.Body.Frame, md)
	}
}

// 从客户端传来的数据中序列化对应过程需要的调用参数
// ok指示数据是否合法
func (s *Server) getCallArgsFromClient(c *websocket.Conn,receiver,method reflect.Value,callerMd,rep *protocol.Message) (callArgs []reflect.Value,ok bool){
	callArgs = []reflect.Value{
		// receiver
		receiver,
	}
	inputTypeList := lreflect.FuncInputTypeList(method)
	for k, v := range callerMd.Body.Frame {
		// 排除receiver
		index := k + 1
		callArg, err := checkCoderType(v, inputTypeList[index])
		if err != nil {
			HandleError(*rep, *ErrServer, c, err.Error())
			return nil,false
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs = append(callArgs, reflect.ValueOf(callArg))
	}
	// 验证客户端传来的栈帧中每个参数的类型是否与服务器需要的一致？
	// receiver(接收器)参与验证
	ok, noMatch := checkInputTypeList(callArgs, inputTypeList)
	if !ok {
		if noMatch != nil {
			HandleError(*rep, *ErrCallArgsType, c,
				fmt.Sprintf("pass value type is %s but call arg type is %s", noMatch[1], noMatch[0]),
			)
		} else {
			HandleError(*rep, *ErrCallArgsType, c,
				fmt.Sprintf("pass arg list length no equal of call arg list : len(callArgs) == %d : len(inputTypeList) == %d",
					len(callArgs)-1, len(inputTypeList)-1),
			)
		}
		return nil,false
	}
	return callArgs,true
}
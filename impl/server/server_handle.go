package server

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/impl/common"
	"github.com/nyan233/littlerpc/impl/internal"
	"github.com/nyan233/littlerpc/impl/transport"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/protocol"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

// try 指示是否需要重入处理结果的逻辑
// cr2 表示内部append过的callResult，以使更改调用者可见
func (s *Server) handleErrAndRepResult(sArg serverCallContext,msg *protocol.Message,callResult []reflect.Value) {

	switch i := lreflect.ToValueTypeEface(callResult[len(callResult)-1]); i.(type) {
	case *protocol.Error:
		// 非默认Codec不注入这些错误，因为像protobuf之类的可能有自己的
		// 编码和Struct Tag规则，这些自动生成的error会导致错误
		if sArg.Codec.Scheme() != "json" {
			break
		}
		err := msg.Encode(sArg.Codec,i)
		if err != nil {
			HandleError(sArg,msg.Header.MsgId,*common.ErrServer, err.Error())
			return
		}
	case error:
		if sArg.Codec.Scheme() != "json" {
			break
		}
		var str = i.(error).Error()
		err := msg.Encode(sArg.Codec,&str)
		if err != nil {
			HandleError(sArg,msg.Header.MsgId,*common.ErrServer, err.Error())
			return
		}
	case nil:
		if sArg.Codec.Scheme() != "json" {
			break
		}
		var intNil int64
		err := msg.Encode(sArg.Codec,&intNil)
		if err != nil {
			HandleError(sArg,msg.Header.MsgId,*common.ErrServer, err.Error())
			return
		}
	default:
		// 不允许最后一个返回值不是*code.Error/error
		panic("last return value no implement error")
	}
	return
}

func (s *Server) sendMsg(sArg serverCallContext,msg *protocol.Message) {
	// TODO : implement text encoding and gzip encoding
	// rep Header已经被调用者提前设置好内容，所以这里发送消息的逻辑不用设置
	// write header
	buf := s.bufferPool.Get().(*transport.BufferPool)
	defer s.bufferPool.Put(buf)
	buf.Buf = buf.Buf[:0]
	buf.Buf = append(buf.Buf,msg.EncodeHeader()...)
	bodyStart := len(buf.Buf)
	// write body
	for _,v := range msg.Body {
		buf.Buf = append(buf.Buf,v...)
	}
	bytes, err := s.encoder.EnPacket(buf.Buf[bodyStart:])
	if err != nil {
		HandleError(sArg,msg.Header.MsgId, *common.ErrServer, err.Error())
		return
	}
	buf.Buf = append(buf.Buf[:bodyStart],bytes...)
	// write data
	_, err = sArg.Conn.Write(buf.Buf)
	if err != nil {
		s.logger.ErrorFromErr(err)
	}
}

// 将用户过程的返回结果集序列化为可传输的json数据
func (s *Server) handleResult(sArg serverCallContext,msg *protocol.Message,callResult []reflect.Value) {
	for _, v := range callResult[:len(callResult)-1] {
		var eface = v.Interface()
		// 可替换的Codec已经不需要Any包装器了
		err := msg.Encode(sArg.Codec,eface)
		if err != nil {
			HandleError(sArg,msg.Header.MsgId, *common.ErrServer, "")
			return
		}
	}
}

// 从客户端传来的数据中序列化对应过程需要的调用参数
// ok指示数据是否合法
func (s *Server) getCallArgsFromClient(sArg serverCallContext,msg *protocol.Message,receiver,method reflect.Value) (callArgs []reflect.Value,ok bool){
	callArgs = []reflect.Value{
		// receiver
		receiver,
	}
	// 排除receiver
	inputTypeList := lreflect.FuncInputTypeList(method,true)
	for k, v := range msg.Body {
		callArg, err := internal.CheckCoderType(sArg.Codec,v, inputTypeList[k])
		if err != nil {
			HandleError(sArg,msg.Header.MsgId, *common.ErrServer, err.Error())
			return nil,false
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs = append(callArgs, reflect.ValueOf(callArg))
	}
	// 验证客户端传来的栈帧中每个参数的类型是否与服务器需要的一致？
	// receiver(接收器)参与验证
	ok, noMatch := internal.CheckInputTypeList(callArgs, append([]interface{}{receiver.Interface()},inputTypeList...))
	if !ok {
		if noMatch != nil {
			HandleError(sArg,msg.Header.MsgId, *common.ErrCallArgsType,
				fmt.Sprintf("pass value type is %s but call arg type is %s", noMatch[1], noMatch[0]),
			)
		} else {
			HandleError(sArg,msg.Header.MsgId, *common.ErrCallArgsType,
				fmt.Sprintf("pass arg list length no equal of call arg list : len(callArgs) == %d : len(inputTypeList) == %d",
					len(callArgs)-1, len(inputTypeList)-1),
			)
		}
		return nil,false
	}
	return callArgs,true
}

func HandleError(sArg serverCallContext,msgId int64, errNo protocol.Error, appendInfo string) {
	codec := protocol.GetCodec("json")
	conn := sArg.Conn
	encoder := packet.GetEncoder("text")
	msg := protocol.Message{}
	// write header
	header := &msg.Header
	header.Timestamp = time.Now().Unix()
	header.MsgType = protocol.MessageReturn
	header.MsgId = msgId
	header.CodecType = codec.Scheme()
	header.Encoding = encoder.Scheme()
	switch errNo.Info {
	case common.ErrJsonUnMarshal.Info, common.ErrMethodNoRegister.Info, common.ErrCallArgsType.Info:
		errNo.Trace += appendInfo
		err := msg.Encode(codec,errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		errNoBytes := common.BufferIoEncodeMessage(&msg)
		errNoBytes, err = encoder.EnPacket(errNoBytes)
		if err != nil {
			panic(errors.New(fmt.Sprintf("encoding/%s enpacket failed",encoder.Scheme())))
		}
		conn.Write(errNoBytes)
		break
	case common.ErrServer.Info:
		errNo.Info += appendInfo
		_, file, line, _ := runtime.Caller(1)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := msg.Encode(codec,errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		errNoBytes := common.BufferIoEncodeMessage(&msg)
		errNoBytes, err = encoder.EnPacket(errNoBytes)
		if err != nil {
			panic(errors.New(fmt.Sprintf("encoding/%s enpacket failed",encoder.Scheme())))
		}
		conn.Write(errNoBytes)
		break
	case common.Nil.Info:
		err := msg.Encode(codec,errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		errNoBytes := common.BufferIoEncodeMessage(&msg)
		errNoBytes, err = encoder.EnPacket(errNoBytes)
		if err != nil {
			panic(errors.New(fmt.Sprintf("encoding/%s enpacket failed",encoder.Scheme())))
		}
		conn.Write(errNoBytes)
	}
}


package server

import (
	"context"
	"fmt"
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/container"
	"github.com/nyan233/littlerpc/protocol"
	perror "github.com/nyan233/littlerpc/protocol/error"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"github.com/nyan233/littlerpc/stream"
	"github.com/nyan233/littlerpc/utils/convert"
	"github.com/nyan233/littlerpc/utils/hash"
	"reflect"
	"strconv"
)

// 必须在其结果集中首先处理错误在处理其余结果
// bool(1)类型的返回值指示用户是否返回了错误,Success在LittleRpc中并不会被认为是错误
// bool(2)类型的返回值指示是否可继续
func (s *Server) handleErrAndRepResult(sArg serverCallContext, msg *protocol.Message, callResult []reflect.Value) (bool, bool) {
	interErr := lreflect.ToValueTypeEface(callResult[len(callResult)-1])
	// 无错误
	if interErr == error(nil) {
		msg.SetMetaData("littlerpc-code", strconv.Itoa(common.Success.Code))
		msg.SetMetaData("littlerpc-message", common.Success.Message)
		return false, true
	}
	// 检查是否实现了自定义错误的接口
	desc, ok := interErr.(perror.LErrorDesc)
	if !ok {
		msg.SetMetaData("littlerpc-code", strconv.Itoa(desc.GetCode()))
		msg.SetMetaData("littlerpc-message", desc.GetMessage())
		bytes, err := desc.MarshalMores()
		if err != nil {
			s.handleError(sArg, msg.MsgId, common.ErrCodecMarshalError,
				fmt.Sprintf("%s : %s", sArg.Codec.Scheme(), err.Error()))
			return true, false
		}
		msg.SetMetaData("littlerpc-mores-bin", convert.BytesToString(bytes))
		return true, true
	}
	err, ok := interErr.(error)
	// NOTE 按理来说, 在正常情况下!ok这个分支不应该被激活, 检查每个过程返回error是Elem的责任
	// NOTE 建立这个分支是防止用户自作聪明使用一些Hack的手段绕过了Elem的检查
	if !ok {
		s.handleError(sArg, msg.MsgId, perror.LNewBaseError(perror.UnsafeOption, "Server.Elem no checker on error"))
		return true, false
	}
	msg.SetMetaData("littlerpc-code", strconv.Itoa(perror.Unknown))
	msg.SetMetaData("littlerpc-message", err.Error())
	return true, true
}

func (s *Server) sendMsg(sArg serverCallContext, msg *protocol.Message) {
	// TODO : implement text encoding and gzip encoding
	// rep Header已经被调用者提前设置好内容，所以这里发送消息的逻辑不用设置
	// write header
	muxMsg := &protocol.MuxBlock{
		Flags:    protocol.MuxEnabled,
		StreamId: hash.FastRand(),
		MsgId:    msg.MsgId,
	}
	bp := sArg.Desc.bytesBuffer.Get().(*container.Slice[byte])
	bp.Reset()
	defer sArg.Desc.bytesBuffer.Put(bp)
	// write body
	if sArg.Encoder.Scheme() != "text" {
		bytes, err := sArg.Encoder.EnPacket(msg.Payloads)
		if err != nil {
			s.handleError(sArg, msg.MsgId, common.ErrServer, err.Error())
			return
		}
		msg.Payloads = append(msg.Payloads[:0], bytes...)
	}
	// 计算真实长度
	msg.PayloadLength = uint32(msg.GetLength())
	protocol.MarshalMessage(msg, bp)
	// write data
	// 大于一个MuxBlock时则分片发送
	sendBuf := sArg.Desc.bytesBuffer.Get().(*container.Slice[byte])
	defer sArg.Desc.bytesBuffer.Put(sendBuf)
	err := common.MuxWriteAll(sArg.Desc, muxMsg, sendBuf, *bp, nil)
	if err != nil {
		s.logger.ErrorFromString(fmt.Sprintf("Marshal MuxBlock failed: %v", sArg.Desc.Close()))
		return
	}
	if err := s.pManager.OnComplete(msg, nil); err != nil {
		s.logger.ErrorFromErr(err)
	}
}

// 将用户过程的返回结果集序列化为可传输的json数据
func (s *Server) handleResult(sArg serverCallContext, msg *protocol.Message, callResult []reflect.Value) {
	for _, v := range callResult[:len(callResult)-1] {
		// NOTE : 对于指针类型或者隐含指针的类型, 他检查用户过程是否返回nil
		// NOTE : 对于非指针的值传递类型, 它检查该类型是否是零值
		// 借助这个哨兵条件可以减少零值的序列化/网络开销
		if v.IsZero() {
			// 添加返回参数的标记, 这是因为在多个返回参数可能出现以下的情况
			// (Value),(Value2),(nil),(Zero)
			// 在以上情况下简单地忽略并不是一个好主意(会导致返回值反序列化异常), 所以需要一个标记让客户端知道
			msg.AppendPayloads(make([]byte, 0))
			continue
		}
		var eface = v.Interface()
		// 可替换的Codec已经不需要Any包装器了
		bytes, err := sArg.Codec.Marshal(eface)
		if err != nil {
			s.handleError(sArg, msg.MsgId, common.ErrServer, "")
			return
		}
		msg.AppendPayloads(bytes)
	}
}

// 从客户端传来的数据中序列化对应过程需要的调用参数
// ok指示数据是否合法
func (s *Server) getCallArgsFromClient(sArg serverCallContext, msg *protocol.Message, receiver, method reflect.Value) (callArgs []reflect.Value, ok bool) {
	callArgs = []reflect.Value{
		// receiver
		receiver,
	}
	// 排除receiver
	iter := msg.PayloadsIterator()
	// 去除接收者之后的输入参数长度
	// 校验客户端传递的参数和服务端是否一致
	if nInput := method.Type().NumIn() - 1; nInput != msg.PayloadLayout.Len() {
		serverErr := *common.ErrServer
		serverErr.Message = "client input args number no equal server"
		s.handleError(sArg, msg.MsgId, &serverErr, fmt.Sprintf("Client : %d", msg.PayloadLayout.Len()),
			fmt.Sprintf("Server : %d", nInput))
		return nil, false
	}
	var inputStart int
	if msg.GetMetaData("context-timeout") != "" {
		inputStart++
	}
	if msg.GetMetaData("stream-id") != "" {
		inputStart++
	}
	inputTypeList := lreflect.FuncInputTypeList(method, inputStart, true, func(i int) bool {
		if msg.PayloadLayout[i] == 0 {
			return true
		}
		return false
	})
	if inputStart == 2 {
		val1 := reflect.New(method.Type().In(1)).Interface()
		val2 := reflect.New(method.Type().In(2)).Interface()
		switch val1.(type) {
		case *context.Context:
			break
		default:
			// TODO Handle Error
		}
		switch val2.(type) {
		case *stream.Stream:
			break
		default:
			// TODO Handle Error
		}
	} else if inputStart == 1 {
		typ1 := method.Type().In(1)
		if typ1.Kind() != reflect.Interface {
			// TODO Handle Error
		}
		switch reflect.New(typ1).Interface().(type) {
		case *context.Context:
			break
		case *stream.Stream:
			break
		default:
			// TODO Handle Error
		}
	}
	for i := 0; i < len(inputTypeList) && iter.Next(); i++ {
		eface := inputTypeList[i]
		argBytes := iter.Take()
		if len(argBytes) == 0 {
			callArgs = append(callArgs, reflect.ValueOf(eface))
			continue
		}
		callArg, err := common.CheckCoderType(sArg.Codec, argBytes, eface)
		if err != nil {
			s.handleError(sArg, msg.MsgId, common.ErrServer, err.Error())
			return nil, false
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs = append(callArgs, reflect.ValueOf(callArg))
	}
	return callArgs, true
}

func (s *Server) handleError(sArg serverCallContext, msgId uint64, errNo *perror.LBaseError, appendInfo ...string) {
	desc := sArg.Desc
	msg := protocol.NewMessage()
	msg.SetMsgType(protocol.MessageReturn)
	msg.MsgId = msgId
	msg.SetMetaData("littlerpc-code", strconv.Itoa(errNo.Code))
	msg.SetMetaData("littlerpc-message", errNo.Message)
	// 为空则不序列化Mores, 否则会造成空间浪费
	if appendInfo != nil || len(appendInfo) > 0 {
		errDesc := s.lNewErrorFn(errNo.Code, errNo.Message, appendInfo)
		bytes, err := errDesc.MarshalMores()
		if err != nil {
			msg.SetMetaData("littlerpc-code", strconv.Itoa(common.ErrServer.Code))
			msg.SetMetaData("littlerpc-message", "Marshal error Mores data failed")
		} else {
			msg.SetMetaData("littlerpc-mores-bin", convert.BytesToString(bytes))
		}
	}
	msg.PayloadLength = uint32(msg.GetLength())
	bp := desc.bytesBuffer.Get().(*container.Slice[byte])
	bp.Reset()
	defer desc.bytesBuffer.Put(bp)
	protocol.MarshalMessage(msg, bp)
	muxBlock := &protocol.MuxBlock{
		Flags:    protocol.MuxEnabled,
		StreamId: hash.FastRand(),
		MsgId:    msgId,
	}
	err := common.MuxWriteAll(desc, muxBlock, nil, *bp, nil)
	if err != nil {
		s.logger.ErrorFromErr(err)
		return
	}
}

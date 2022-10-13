package server

import (
	"context"
	"fmt"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
	common2 "github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/stream"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	"github.com/nyan233/littlerpc/pkg/utils/message"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"reflect"
	"strconv"
	"sync"
)

// 必须在其结果集中首先处理错误在处理其余结果
func (s *Server) setErrResult(msg *protocol.Message, callResult reflect.Value) perror.LErrorDesc {
	interErr := reflect2.ToValueTypeEface(callResult)
	// 无错误
	if interErr == error(nil) {
		msg.SetMetaData("littlerpc-code", strconv.Itoa(common2.Success.Code()))
		msg.SetMetaData("littlerpc-message", common2.Success.Message())
		return nil
	}
	// 检查是否实现了自定义错误的接口
	desc, ok := interErr.(perror.LErrorDesc)
	if ok {
		msg.SetMetaData("littlerpc-code", strconv.Itoa(desc.Code()))
		msg.SetMetaData("littlerpc-message", desc.Message())
		bytes, err := desc.MarshalMores()
		if err != nil {
			return s.eHandle.LWarpErrorDesc(
				common2.ErrCodecMarshalError,
				fmt.Sprintf("%s : %s", "littlerpc-mores-bin", err.Error()))
		}
		msg.SetMetaData("littlerpc-mores-bin", convert.BytesToString(bytes))
		return nil
	}
	err, ok := interErr.(error)
	// NOTE 按理来说, 在正常情况下!ok这个分支不应该被激活, 检查每个过程返回error是Elem的责任
	// NOTE 建立这个分支是防止用户自作聪明使用一些Hack的手段绕过了Elem的检查
	if !ok {
		return s.eHandle.LNewErrorDesc(perror.UnsafeOption, "Server.Elem no checker on error")
	}
	msg.SetMetaData("littlerpc-code", strconv.Itoa(perror.Unknown))
	msg.SetMetaData("littlerpc-message", err.Error())
	return nil
}

func (s *Server) processAndSendMsg(msgOpt *messageOpt, msg *protocol.Message, useMux bool) {
	// TODO : implement text encoding and gzip encoding
	// rep Header已经被调用者提前设置好内容，所以这里发送消息的逻辑不用设置
	// write header
	bytesBuffer := sharedPool.TakeBytesPool()
	bp := bytesBuffer.Get().(*container.Slice[byte])
	bp.Reset()
	defer bytesBuffer.Put(bp)
	// write body
	if msgOpt.Encoder.Scheme() != "text" {
		bytes, err := msgOpt.Encoder.EnPacket(msg.Payloads)
		if err != nil {
			s.handleError(msgOpt.Conn, msg.MsgId, s.eHandle.LWarpErrorDesc(common2.ErrServer, err.Error()))
			return
		}
		msg.Payloads = append(msg.Payloads[:0], bytes...)
	}
	// 计算真实长度
	msg.PayloadLength = uint32(msg.GetLength())
	protocol.MarshalMessage(msg, bp)
	// 不使用Mux消息的情况
	if !useMux {
		s.sendOnNoMux(msgOpt, msg, *bp)
	} else {
		s.sendOnMux(msgOpt, bytesBuffer, msg, *bp)
	}
}

func (s *Server) sendOnNoMux(msgOpt *messageOpt, msg *protocol.Message, bytes []byte) {
	err := common2.WriteControl(msgOpt.Conn, bytes)
	if err != nil {
		s.logger.ErrorFromString(fmt.Sprintf("Write NoMuxMessage failed: %v", msgOpt.Conn.Close()))
	}
	if s.debug {
		s.logger.Debug(message.AnalysisMessage(bytes).String())
	}
	if err := s.pManager.OnComplete(msg, err); err != nil {
		s.logger.ErrorFromErr(err)
	}
	return
}

func (s *Server) sendOnMux(msgOpt *messageOpt, bytesBuffer *sync.Pool, msg *protocol.Message, bytes []byte) {
	muxMsg := &protocol.MuxBlock{
		Flags:    protocol.MuxEnabled,
		StreamId: random.FastRand(),
		MsgId:    msg.MsgId,
	}
	// write data
	// 大于一个MuxBlock时则分片发送
	sendBuf := bytesBuffer.Get().(*container.Slice[byte])
	defer bytesBuffer.Put(sendBuf)
	err := common2.MuxWriteAll(msgOpt.Conn, muxMsg, sendBuf, bytes, nil)
	if err != nil {
		s.logger.ErrorFromString(fmt.Sprintf("Write MuxMessage failed: %v", msgOpt.Conn.Close()))
		return
	}
	if err := s.pManager.OnComplete(msg, err); err != nil {
		s.logger.ErrorFromErr(err)
	}
}

// 将用户过程的返回结果集序列化为可传输的json数据
func (s *Server) handleResult(msgOpt *messageOpt, msg *protocol.Message, callResult []reflect.Value) {
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
		bytes, err := msgOpt.Codec.Marshal(eface)
		if err != nil {
			s.handleError(msgOpt.Conn, msg.MsgId, common2.ErrServer)
			return
		}
		msg.AppendPayloads(bytes)
	}
}

// 从客户端传来的数据中序列化对应过程需要的调用参数
// ok指示数据是否合法
func (s *Server) getCallArgsFromClient(sArg serverCallContext, msg *protocol.Message, receiver, method reflect.Value) ([]reflect.Value, perror.LErrorDesc) {
	callArgs := []reflect.Value{
		// receiver
		receiver,
	}
	// 排除receiver
	iter := msg.PayloadsIterator()
	// 去除接收者之后的输入参数长度
	// 校验客户端传递的参数和服务端是否一致
	if nInput := method.Type().NumIn() - 1; nInput != msg.PayloadLayout.Len() {
		return nil, s.eHandle.LWarpErrorDesc(common2.ErrServer,
			"client input args number no equal server",
			fmt.Sprintf("Client : %d", msg.PayloadLayout.Len()), fmt.Sprintf("Server : %d", nInput))
	}
	var inputStart int
	if msg.GetMetaData("context-timeout") != "" {
		inputStart++
	}
	if msg.GetMetaData("stream-id") != "" {
		inputStart++
	}
	inputTypeList := reflect2.FuncInputTypeList(method, inputStart, true, func(i int) bool {
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
		case *stream.LStream:
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
		case *stream.LStream:
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
		callArg, err := common2.CheckCoderType(sArg.Codec, argBytes, eface)
		if err != nil {
			return nil, s.eHandle.LWarpErrorDesc(common2.ErrServer, err.Error())
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs = append(callArgs, reflect.ValueOf(callArg))
	}
	return callArgs, nil
}

// NOTE: 这里负责处理Server遇到的所有错误, 它会在遇到严重的错误时关闭连接, 不那么重要的错误则尝试返回给客户端
// NOTE: 严重错误 -> UnsafeOption | MessageDecodingFailed | MessageEncodingFailed
// NOTE: 轻微错误 -> 除了严重错误都是
// Update: LittleRpc现在的错误返回统一使用NoMux类型的消息
func (s *Server) handleError(desc transport.ConnAdapter, msgId uint64, errNo perror.LErrorDesc) {
	bytesBuffer := sharedPool.TakeBytesPool()
	switch errNo.Code() {
	case perror.UnsafeOption, perror.MessageDecodingFailed, perror.MessageEncodingFailed:
		// 严重影响到后续运行的错误需要关闭连接
		s.logger.ErrorFromErr(errNo)
		err := desc.Close()
		if err != nil {
			s.logger.ErrorFromErr(err)
		}
	default:
		// 普通一些的错误可以不关闭连接
		msg := protocol.NewMessage()
		msg.SetMsgType(protocol.MessageReturn)
		msg.MsgId = msgId
		msg.SetMetaData("littlerpc-code", strconv.Itoa(errNo.Code()))
		msg.SetMetaData("littlerpc-message", errNo.Message())
		// 为空则不序列化Mores, 否则会造成空间浪费
		mores := errNo.Mores()
		if mores != nil && len(mores) > 0 {
			bytes, err := errNo.MarshalMores()
			if err != nil {
				s.logger.ErrorFromErr(err)
				_ = desc.Close()
				return
			} else {
				msg.SetMetaData("littlerpc-mores-bin", convert.BytesToString(bytes))
			}
		}
		msg.PayloadLength = uint32(msg.GetLength())
		bp := bytesBuffer.Get().(*container.Slice[byte])
		bp.Reset()
		defer bytesBuffer.Put(bp)
		protocol.MarshalMessage(msg, bp)
		err := common2.WriteControl(desc, *bp)
		if err != nil {
			s.logger.ErrorFromErr(err)
			_ = desc.Close()
			return
		}
	}
}

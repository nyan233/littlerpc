package server

import (
	"fmt"
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/container"
	"github.com/nyan233/littlerpc/protocol"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"github.com/nyan233/littlerpc/utils/hash"
	"reflect"
)

// try 指示是否需要重入处理结果的逻辑
// cr2 表示内部append过的callResult，以使更改调用者可见
func (s *Server) handleErrAndRepResult(sArg serverCallContext, msg *protocol.Message, callResult []reflect.Value) {
	bytes, err := sArg.Codec.MarshalError(lreflect.ToValueTypeEface(callResult[len(callResult)-1]))
	if err != nil {
		s.handleError(sArg, msg.MsgId, *common.ErrCodecMarshalError,
			fmt.Sprintf("%s : %s", sArg.Codec.Scheme(), err.Error()))
		return
	}
	msg.AppendPayloads(bytes)
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
			s.handleError(sArg, msg.MsgId, *common.ErrServer, err.Error())
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
		var eface = v.Interface()
		// 可替换的Codec已经不需要Any包装器了
		bytes, err := sArg.Codec.Marshal(eface)
		if err != nil {
			s.handleError(sArg, msg.MsgId, *common.ErrServer, "")
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
	inputTypeList := lreflect.FuncInputTypeList(method, true)
	var i int
	var e error
	protocol.RangePayloads(msg, msg.Payloads, func(p []byte, endBefore bool) bool {
		eface := inputTypeList[i]
		callArg, err := common.CheckCoderType(sArg.Codec, p, eface)
		if err != nil {
			s.handleError(sArg, msg.MsgId, *common.ErrServer, err.Error())
			e = err
			return false
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs = append(callArgs, reflect.ValueOf(callArg))
		i++
		return true
	})
	if e != nil {
		return nil, false
	}
	// 验证客户端传来的栈帧中每个参数的类型是否与服务器需要的一致？
	// receiver(接收器)参与验证
	ok, noMatch := common.CheckInputTypeList(callArgs, append([]interface{}{receiver.Interface()}, inputTypeList...))
	if !ok {
		if noMatch != nil {
			s.handleError(sArg, msg.MsgId, *common.ErrCallArgsType,
				fmt.Sprintf("pass value type is %s but call arg type is %s", noMatch[1], noMatch[0]),
			)
		} else {
			s.handleError(sArg, msg.MsgId, *common.ErrCallArgsType,
				fmt.Sprintf("pass arg list length no equal of call arg list : len(callArgs) == %d : len(inputTypeList) == %d",
					len(callArgs)-1, len(inputTypeList)-1),
			)
		}
		return nil, false
	}
	return callArgs, true
}

// TODO 需要修改兼容Mux的逻辑
func (s *Server) handleError(sArg serverCallContext, msgId uint64, errNo protocol.Error, appendInfo string) {
	desc := sArg.Desc
	msg := protocol.NewMessage()
	// 表示该消息类型是服务器的错误返回
	msg.SetMsgType(protocol.MessageErrorReturn)
	msg.MsgId = msgId
	// 设置error
	errNo.Trace = appendInfo
	msg.Payloads = []byte(errNo.Error())
	bp := desc.bytesBuffer.Get().(*container.Slice[byte])
	bp.Reset()
	defer desc.bytesBuffer.Put(bp)
	protocol.MarshalMessage(msg, bp)
	desc.Lock()
	defer desc.Unlock()
	_, err := desc.Write(*bp)
	if err != nil {
		s.logger.ErrorFromErr(err)
		return
	}
}

package server

import (
	"fmt"
	"github.com/nyan233/littlerpc/impl/common"
	"github.com/nyan233/littlerpc/impl/internal"
	"github.com/nyan233/littlerpc/protocol"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"reflect"
	"time"
)

// try 指示是否需要重入处理结果的逻辑
// cr2 表示内部append过的callResult，以使更改调用者可见
func (s *Server) handleErrAndRepResult(sArg serverCallContext, msg *protocol.Message, callResult []reflect.Value) {
	bytes, err := sArg.Codec.MarshalError(lreflect.ToValueTypeEface(callResult[len(callResult)-1]))
	if err != nil {
		s.HandleError(sArg, msg.Header.MsgId, *common.ErrCodecMarshalError,
			fmt.Sprintf("%s : %s", sArg.Codec.Scheme(), err.Error()))
		return
	}
	msg.DecodeRaw(bytes)
}

func (s *Server) sendMsg(sArg serverCallContext, msg *protocol.Message) {
	// TODO : implement text encoding and gzip encoding
	// rep Header已经被调用者提前设置好内容，所以这里发送消息的逻辑不用设置
	// write header
	bp := s.bufferPool.Get().(*[]byte)
	*bp = (*bp)[:0]
	defer s.bufferPool.Put(bp)
	msg.EncodeHeaderFormBufferPool(bp)
	bodyStart := len(*bp)
	// write body
	msg.EncodeBodyFormBufferPool(bp)
	bytes, err := sArg.Encoder.EnPacket((*bp)[bodyStart:])
	if err != nil {
		s.HandleError(sArg, msg.Header.MsgId, *common.ErrServer, err.Error())
		return
	}
	*bp = append((*bp)[:bodyStart], bytes...)
	// write data
	_, err = sArg.Conn.Write(*bp)
	if err != nil {
		s.logger.ErrorFromErr(err)
	}
}

// 将用户过程的返回结果集序列化为可传输的json数据
func (s *Server) handleResult(sArg serverCallContext, msg *protocol.Message, callResult []reflect.Value) {
	for _, v := range callResult[:len(callResult)-1] {
		var eface = v.Interface()
		// 可替换的Codec已经不需要Any包装器了
		err := msg.Encode(sArg.Codec, eface)
		if err != nil {
			s.HandleError(sArg, msg.Header.MsgId, *common.ErrServer, "")
			return
		}
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
	for k, v := range msg.Body {
		eface := inputTypeList[k]
		callArg, err := internal.CheckCoderType(sArg.Codec, v, eface)
		if err != nil {
			s.HandleError(sArg, msg.Header.MsgId, *common.ErrServer, err.Error())
			return nil, false
		}
		// 可以根据获取的参数类别的每一个参数的类型信息得到
		// 所需的精确类型，所以不用再对变长的类型做处理
		callArgs = append(callArgs, reflect.ValueOf(callArg))
	}
	// 验证客户端传来的栈帧中每个参数的类型是否与服务器需要的一致？
	// receiver(接收器)参与验证
	ok, noMatch := internal.CheckInputTypeList(callArgs, append([]interface{}{receiver.Interface()}, inputTypeList...))
	if !ok {
		if noMatch != nil {
			s.HandleError(sArg, msg.Header.MsgId, *common.ErrCallArgsType,
				fmt.Sprintf("pass value type is %s but call arg type is %s", noMatch[1], noMatch[0]),
			)
		} else {
			s.HandleError(sArg, msg.Header.MsgId, *common.ErrCallArgsType,
				fmt.Sprintf("pass arg list length no equal of call arg list : len(callArgs) == %d : len(inputTypeList) == %d",
					len(callArgs)-1, len(inputTypeList)-1),
			)
		}
		return nil, false
	}
	return callArgs, true
}

func (s *Server) HandleError(sArg serverCallContext, msgId int64, errNo protocol.Error, appendInfo string) {
	conn := sArg.Conn
	msg := protocol.Message{}
	// write header
	header := &msg.Header
	header.Timestamp = time.Now().Unix()
	// 表示该消息类型是服务器的错误返回
	header.MsgType = protocol.MessageErrorReturn
	header.MsgId = msgId
	// 设置error
	errNo.Trace = appendInfo
	msg.DecodeRaw([]byte(errNo.Error()))
	bp := msg.EncodeHeaderAndBodyFromBufferPool(&s.bufferPool)
	defer s.bufferPool.Put(bp)
	_, err := conn.Write(*bp)
	if err != nil {
		s.logger.ErrorFromErr(err)
		return
	}
}

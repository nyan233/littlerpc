package server

import (
	"fmt"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
	common2 "github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/common/utils/debug"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/control"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	perror "github.com/nyan233/littlerpc/protocol/error"
	"github.com/nyan233/littlerpc/protocol/message"
	"reflect"
	"strconv"
)

// 必须在其结果集中首先处理错误在处理其余结果
func (s *Server) setErrResult(msg *message.Message, callResult reflect.Value) perror.LErrorDesc {
	interErr := reflect2.ToValueTypeEface(callResult)
	// 无错误
	if interErr == error(nil) {
		msg.MetaData.Store(message.ErrorCode, strconv.Itoa(common2.Success.Code()))
		msg.MetaData.Store(message.ErrorMessage, common2.Success.Message())
		return nil
	}
	// 检查是否实现了自定义错误的接口
	desc, ok := interErr.(perror.LErrorDesc)
	if ok {
		msg.MetaData.Store(message.ErrorCode, strconv.Itoa(desc.Code()))
		msg.MetaData.Store(message.ErrorMessage, desc.Message())
		bytes, err := desc.MarshalMores()
		if err != nil {
			return s.eHandle.LWarpErrorDesc(
				common2.ErrCodecMarshalError,
				fmt.Sprintf("%s : %s", message.ErrorMore, err.Error()))
		}
		msg.MetaData.Store(message.ErrorMore, convert.BytesToString(bytes))
		return nil
	}
	err, ok := interErr.(error)
	// NOTE 按理来说, 在正常情况下!ok这个分支不应该被激活, 检查每个过程返回error是Elem的责任
	// NOTE 建立这个分支是防止用户自作聪明使用一些Hack的手段绕过了Elem的检查
	if !ok {
		return s.eHandle.LNewErrorDesc(perror.UnsafeOption, "Server.RegisterClass no checker on error")
	}
	msg.MetaData.Store(message.ErrorCode, strconv.Itoa(perror.Unknown))
	msg.MetaData.Store(message.ErrorMessage, err.Error())
	return nil
}

func (s *Server) encodeAndSendMsg(msgOpt messageOpt, msg *message.Message, useMux bool) {
	err := msgOpt.Writer.Writer(msgwriter.Argument{
		Message: msg,
		Conn:    msgOpt.Conn,
		Option:  &metadata.ProcessOption{UseMux: useMux},
		Encoder: msgOpt.Encoder,
		Pool:    sharedPool.TakeBytesPool(),
		OnDebug: debug.MessageDebug(s.logger, s.debug, useMux),
		EHandle: s.eHandle,
	})
	if err != nil {
		pErr := s.pManager.OnComplete(msg, err)
		if err != nil {
			s.logger.ErrorFromErr(pErr)
		}
		s.handleError(msgOpt.Conn, msg.GetMsgId(), err)
	}
}

// 将用户过程的返回结果集序列化为可传输的json数据
func (s *Server) handleResult(msgOpt messageOpt, msg *message.Message, callResult []reflect.Value) {
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
			s.handleError(msgOpt.Conn, msg.GetMsgId(), common2.ErrServer)
			return
		}
		msg.AppendPayloads(bytes)
	}
}

// NOTE: 这里负责处理Server遇到的所有错误, 它会在遇到严重的错误时关闭连接, 不那么重要的错误则尝试返回给客户端
// NOTE: 严重错误 -> UnsafeOption | MessageDecodingFailed | MessageEncodingFailed
// NOTE: 轻微错误 -> 除了严重错误都是
// Update: LittleRpc现在的错误返回统一使用NoMux类型的消息
func (s *Server) handleError(desc transport.ConnAdapter, msgId uint64, errNo perror.LErrorDesc) {
	bytesBuffer := sharedPool.TakeBytesPool()
	switch errNo.Code() {
	case perror.ConnectionErr:
		// 连接错误默认已经被关闭连接, 所以打印日志即可
		s.logger.ErrorFromErr(errNo)
	case perror.UnsafeOption, perror.MessageDecodingFailed, perror.MessageEncodingFailed:
		// 严重影响到后续运行的错误需要关闭连接
		s.logger.ErrorFromErr(errNo)
		err := desc.Close()
		if err != nil {
			s.logger.ErrorFromErr(err)
		}
	default:
		// 普通一些的错误可以不关闭连接
		msg := message.New()
		msg.SetMsgType(message.Return)
		msg.SetMsgId(msgId)
		msg.MetaData.Store(message.ErrorCode, strconv.Itoa(errNo.Code()))
		msg.MetaData.Store(message.ErrorMessage, errNo.Message())
		// 为空则不序列化Mores, 否则会造成空间浪费
		mores := errNo.Mores()
		if mores != nil && len(mores) > 0 {
			bytes, err := errNo.MarshalMores()
			if err != nil {
				s.logger.ErrorFromErr(err)
				_ = desc.Close()
				return
			} else {
				msg.MetaData.Store(message.ErrorMore, convert.BytesToString(bytes))
			}
		}
		bp := bytesBuffer.Get().(*container.Slice[byte])
		bp.Reset()
		defer bytesBuffer.Put(bp)
		message.Marshal(msg, bp)
		err := control.WriteControl(desc, *bp)
		if err != nil {
			s.logger.ErrorFromErr(err)
			_ = desc.Close()
			return
		}
	}
}

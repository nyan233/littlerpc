package msgparser

import (
	"github.com/nyan233/littlerpc/pkg/common/jsonrpc2"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/message/mux"
	"math"
)

var handlerCollect [math.MaxUint8]MessageHandler

type Action int
type BaseLenType int

const (
	UnmarshalBase     Action = 0x34 // 已经序列化基本信息, 但是还够不成一个完整的消息, 需要将消息提升到noReadyBuffer中
	UnmarshalComplete Action = 0x45 // 序列化完整消息完成

	// SingleRequest 数据在单次请求中被传完, 适合HTTP之类的协议, LMessageParser在遇到
	// 这个选项时会直接之间使用ParseMsg()传入的bytes来触发Unmarshal()
	// TCP之类的协议使用这个选项会引发半包之类的问题
	SingleRequest BaseLenType = 0x65
	// MultiRequest 数据在可能在多次请求中被传完, 适合TCP之类的协议
	MultiRequest BaseLenType = 0x76
)

type MessageHandler interface {
	BaseLen() (BaseLenType, int)
	MessageLength(base []byte) int
	Unmarshal(data []byte, msg *message.Message) (Action, error)
}

func RegisterMessageHandler(magicNumber uint8, handler MessageHandler) {
	handlerCollect[magicNumber] = handler
}

func GetMessageHandler(magicNumber uint8) MessageHandler {
	return handlerCollect[magicNumber]
}

func init() {
	RegisterMessageHandler(message.MagicNumber, &noMuxHandler{})
	RegisterMessageHandler(mux.Enabled, &muxHandler{})
	RegisterMessageHandler(jsonrpc2.Header, &JsonRpc2Handler{Codec: codec.GetCodec("json").Instance()})
}

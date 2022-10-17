package msgparser

import (
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
	"math"
)

var handlerCollect [math.MaxUint8]MessageHandler

type Action int

const (
	UnmarshalBase     Action = 0x34 // 已经序列化基本信息, 但是还够不成一个完整的消息, 需要将消息提升到noReadyBuffer中
	UnmarshalComplete Action = 0x45 // 序列化完整消息完成
)

type MessageHandler interface {
	BaseLen() int
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
	RegisterMessageHandler(mux.MuxEnabled, &muxHandler{})
}

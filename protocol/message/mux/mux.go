package mux

import (
	"github.com/nyan233/littlerpc/pkg/container"
)

const (
	// Enabled 指示在Stream中开启Mux的功能
	Enabled uint8 = 0b01
)

const (
	// BlockBaseLen 基本长度, 不包含载荷数据, 因为载荷数据的长度可变
	// - (3 + 6) 是因为考虑到了Struct的对齐
	BlockBaseLen = 1 + 4 + 8 + 2
	// MaxBlockSize Mux模式下Server一次接收多少长度的消息
	MaxBlockSize = 1400
	// MaxPayloadSizeOnMux 在Mux模式上发送的消息一次最多携带多少属于Message的数据
	MaxPayloadSizeOnMux = MaxBlockSize - BlockBaseLen
)

// Block 对一次Mux中的消息帧描述
// StreamId必须确保Server和多个Client之间唯一
type Block struct {
	// Flags可以标记是否开启了Mux
	Flags uint8
	// 通过连接的流的Id
	// 在开启Mux的时候一个连接中可能会有多个流的数据
	StreamId uint32
	// 通过此流传输的消息的Id
	MsgId uint64
	// 消息的载荷数据长度
	PayloadLength uint16
	// 消息的载荷数据
	Payloads container.Slice[byte]
}

// Reset implement Reset interface
func (m *Block) Reset() {
	p := m.Payloads
	p.Reset()
	*m = Block{}
	m.Payloads = p
}

func (m *Block) GetFlags() uint8 {
	return m.Flags
}

func (m *Block) GetStreamId() uint32 {
	return m.StreamId
}

func (m *Block) GetPayloadLength() uint16 {
	return m.PayloadLength
}

func (m *Block) GetMsgId() uint64 {
	return m.MsgId
}

func (m *Block) GetPayloads() []byte {
	return m.Payloads
}

func (m *Block) SetFlags(flags uint8) {
	m.Flags = flags
}

func (m *Block) SetStreamId(streamId uint32) {
	m.StreamId = streamId
}

func (m *Block) SetMsgId(msgId uint64) {
	m.MsgId = msgId
}

func (m *Block) SetPayloads(payloads []byte) {
	m.PayloadLength = uint16(len(payloads))
	m.Payloads = payloads
}

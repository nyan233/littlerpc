package protocol

import (
	"github.com/nyan233/littlerpc/pkg/container"
)

const (
	// MuxEnabled 指示在Stream中开启Mux的功能
	MuxEnabled uint8 = 0b01
)

const (
	// MuxBlockBaseLen 基本长度, 不包含载荷数据, 因为载荷数据的长度可变
	// - (3 + 6) 是因为考虑到了Struct的对齐
	MuxBlockBaseLen = 1 + 4 + 8 + 2
	// MuxMessageBlockSize Mux模式下Server一次接收多少长度的消息
	MuxMessageBlockSize = 1400
	// MaxPayloadSizeOnMux 在Mux模式上发送的消息一次最多携带多少属于Message的数据
	MaxPayloadSizeOnMux = MuxMessageBlockSize - MuxBlockBaseLen
)

// MuxBlock 对一次Mux中的消息帧描述
// StreamId必须确保Server和多个Client之间唯一
type MuxBlock struct {
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
func (m *MuxBlock) Reset() {
	p := m.Payloads
	p.Reset()
	*m = MuxBlock{}
	m.Payloads = p
}

func (m *MuxBlock) GetFlags() uint8 {
	return m.Flags
}

func (m *MuxBlock) GetStreamId() uint32 {
	return m.StreamId
}

func (m *MuxBlock) GetPayloadLength() uint16 {
	return m.PayloadLength
}

func (m *MuxBlock) GetMsgId() uint64 {
	return m.MsgId
}

func (m *MuxBlock) GetPayloads() []byte {
	return m.Payloads
}

func (m *MuxBlock) SetFlags(flags uint8) {
	m.Flags = flags
}

func (m *MuxBlock) SetStreamId(streamId uint32) {
	m.StreamId = streamId
}

func (m *MuxBlock) SetMsgId(msgId uint64) {
	m.MsgId = msgId
}

func (m *MuxBlock) SetPayloads(payloads []byte) {
	m.PayloadLength = uint16(len(payloads))
	m.Payloads = payloads
}

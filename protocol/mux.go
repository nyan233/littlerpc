package protocol

import (
	"github.com/nyan233/littlerpc/container"
)

const (
	// MuxEnabled 指示在Stream中开启Mux的功能
	MuxEnabled uint8 = 0b01
	// MuxDisabled 指示在Stream中关闭Mux的功能
	MuxDisabled uint8 = 0b11
	// OnComplete 指示消息接收完成,这个标志被设置证明这是MsgId对应的数据的最后一段载荷
	// 在Mux关闭时,它应该总是被设置,只有这样LittleRpc才能读取完后续的载荷数据
	OnComplete uint8 = 0b00001
	// NoComplete 指示消息接收未完成,当MsgId对应的数据还存在未发送完的载荷数据时
	// Client应该去设置这个值,否则LittleRpc将丢弃后续的数据,未发送完的数据可能会被解析失败并返回错误
	NoComplete uint8 = 0b00011
)

const (
	// MuxBlockBaseLen 基本长度, 不包含载荷数据, 因为载荷数据的长度可变
	// - (3 + 6) 是因为考虑到了Struct的对齐
	MuxBlockBaseLen = 1 + 4 + 8 + 2
	// MuxMessageBlockSize Mux模式下Server一次接收多少长度的消息
	MuxMessageBlockSize = 4096
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

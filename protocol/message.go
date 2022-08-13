package protocol

import (
	"github.com/nyan233/littlerpc/container"
	"unsafe"
)

const (
	MagicNumber uint8 = 0x45
	// MessageCall 表示这是一条未开启Mux的同步调用的消息
	MessageCall uint8 = 0x10
	// MessageReturn 表示这是一条未开启Mux的同步调用返回消息
	MessageReturn uint8 = 0x18
	// MessageMuxCall 表示这是一条开启Mux的同步调用消息
	MessageMuxCall uint8 = 0x50
	// MessageMuxReturn 表示这是一条开启Mux的同步调用返回消息
	MessageMuxReturn uint8 = 0x56
	// MessageContextWith 用于服务端接收的context.Context初始化
	MessageContextWith uint8 = 0x22
	// MessageContextCancel 用户服务端接收的context.Context的取消API
	MessageContextCancel uint8 = 0x24
	// MessageErrorReturn 调用因某种原因失败的返回
	MessageErrorReturn uint8 = 0x26
	// MessagePing Ping消息
	MessagePing uint8 = 0x33
	// MessagePong Pong消息
	MessagePong uint8 = 0x35

	// MuxMessageBlockSize Mux模式下Server一次接收多少长度的消息
	MuxMessageBlockSize = 4096

	DefaultEncodingType uint8 = 0 // text == json
	DefaultCodecType    uint8 = 0 // encoding == text
)

var (
	FourBytesPadding  = []byte{0, 0, 0, 0}
	EightBytesPadding = []byte{0, 0, 0, 0, 0, 0, 0, 0}
)

type Reset interface {
	Reset()
}

func NewMessage() *Message {
	return &Message{
		MetaData:      container.NewSliceMap[string, string](4),
		PayloadLayout: make([]uint32, 0, 2),
		Payloads:      nil,
	}
}

//	Message 是对一次RPC调用传递的数据的描述
//	封装的方法均不是线程安全的
//	DecodeHeader会修改一些内部值，调用时需要注意顺序
//	为了使用一致性的API，在访问内部一些简单的属性时，请使用Getxx方法
//	在设置一些值的过程中可能需要调整其它值，所以请使用Setxx方法
type Message struct {
	// int/uint数值统一使用大端序
	//	[0] == Magic (魔数，表示这是由littlerpc客户端发起或者服务端回复)
	//	[1] == MsgType (call/return & ping/pong)
	//	[2] == Encoding (default text/gzip)
	//	[3] == CodecType (default json)
	Scope [4]uint8
	// 消息ID，用于表示一次完整的call/return的回复
	MsgId uint64
	// 载荷数据的总长度
	PayloadLength uint32
	// 实例名和调用方法名的布局
	//	InstanceName-Size|MethodName-Size
	NameLayout [2]uint32
	// 实例名
	InstanceName string
	// 要调用的方法名
	MethodName string
	// NOTE:
	//	有效载荷和元数据的范围
	//	在Mux模式中MetaData及其Desc必须能在一个MuxBlock下被装下，否则将会丢失数据
	// 元数据的布局
	//	NMetaData(4 Byte)|Key-Size(4 Byte)|Value-Size(4 Byte)|Key|Size
	// Example :
	//	"hello":"world","world:hello"
	// OutPut:
	//	0x00000002|0x00000005|0x00000005|hello|world|0x00000005|0x00000005|world|hello
	MetaData *container.SliceMap[string, string]
	// 有效载荷数据的布局描述
	// Format :
	//	NArgs(4 Byte)|Arg1-Size(4 Byte)|Arg2-Size(4 Byte)|Arg3-Size(4 Byte)
	// Example :
	//	{"mypyload1":"haha"},{"mypyload2":"hehe"}
	// OutPut:
	//	0x00000002|0x00000014|0x00000014
	PayloadLayout container.Slice[uint32]
	// 调用参数序列化后的载荷数据
	// 如果被压缩了那么在反序列化时,最后剩下的数据均为参数载荷
	Payloads container.Slice[byte]
}

// BaseLength 获取基本数据的长度防止输入过短的数据导致panic
func (m *Message) BaseLength() int {
	ml := m.MinMux()
	// MinMux + NameLayout + NMetaData + NArgs
	return ml + (2 * 4) + (4 * 2)
}

// GetLength 根据结构计算序列化之后的数据长度
func (m *Message) GetLength() int {
	// Scope
	baseLen := len(m.Scope)
	// MsgId & PayloadLength
	baseLen += 12
	// NameLayout
	baseLen += len(m.NameLayout) * 4
	// InstanceName & MethodName
	baseLen += int(m.NameLayout[0] + m.NameLayout[1])
	if m.MetaData != nil && m.MetaData.Len() > 0 {
		// NMetaData
		baseLen += 4
		// Key&Value Struct MetaData
		baseLen += (m.MetaData.Len() * 4) * 2
		// Key & Value Size
		m.MetaData.Range(func(k, v string) bool {
			baseLen += len(k) + len(v)
			return true
		})
	}
	if m.PayloadLayout != nil && m.PayloadLayout.Len() > 0 {
		// NArgs
		baseLen += 4
		// PayloadLayout
		baseLen += len(m.PayloadLayout) * 4
		// Payloads
		baseLen += len(m.Payloads)
	}
	return baseLen
}

func (m *Message) GetCodecType() uint8 {
	return m.Scope[3]
}

func (m *Message) GetEncoderType() uint8 {
	return m.Scope[2]
}

func (m *Message) GetMsgType() uint8 {
	return m.Scope[1]
}

func (m *Message) GetInstanceName() string {
	return m.InstanceName
}

func (m *Message) GetMethodName() string {
	return m.MethodName
}

func (m *Message) GetMetaData(key string) string {
	v, _ := m.MetaData.Load(key)
	return v
}

func (m *Message) SetMetaData(key, value string) {
	m.MetaData.Store(key, value)
}

func (m *Message) RangeMetaData(fn func(key, value string) bool) {
	m.MetaData.Range(fn)
}

func (m *Message) SetMsgId(id uint64) {
	m.MsgId = id
}

func (m *Message) SetCodecType(codecType uint8) {
	m.Scope[3] = codecType
}

func (m *Message) SetEncoderType(encoderTyp uint8) {
	m.Scope[2] = encoderTyp
}

func (m *Message) SetMsgType(msgTyp uint8) {
	m.Scope[1] = msgTyp
}

func (m *Message) SetInstanceName(instanceName string) {
	m.NameLayout[0] = uint32(len(instanceName))
	m.InstanceName = instanceName
}

func (m *Message) SetMethodName(methodName string) {
	m.NameLayout[1] = uint32(len(methodName))
	m.MethodName = methodName
}

func (m *Message) AppendPayloads(p []byte) {
	m.Payloads = append(m.Payloads, p...)
	m.PayloadLayout = append(m.PayloadLayout, uint32(len(p)))
}

func (m *Message) MinMux() int {
	// Scope + MsgId + PayloadLength
	return len(m.Scope) + 8 + 4
}

// Reset 给内存复用的操作提供一致性的语义
func (m *Message) Reset() {
	m.PayloadLayout = nil
	m.Payloads = nil
	*(*uint32)(unsafe.Pointer(&m.Scope)) = 0
	m.InstanceName = ""
	m.MethodName = ""
	m.MsgId = 0
	m.MetaData.Reset()
	m.PayloadLayout.Reset()
	m.Payloads.Reset()
}

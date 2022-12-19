package message

import (
	container2 "github.com/nyan233/littlerpc/core/container"
)

const (
	MagicNumber uint8 = 0x45
	// Call 表示这是一条调用的消息
	Call uint8 = 0x10
	// Return 表示这是一条调用返回消息
	Return uint8 = 0x18
	// ContextCancel 用户服务端接收的context.Context的取消API
	ContextCancel uint8 = 0x24
	// Ping Ping消息
	Ping uint8 = 0x33
	// Pong Pong消息
	Pong uint8 = 0x35

	// BaseLen 的基本长度
	BaseLen = _ScopeLength + 4 + 8
	// DefaultPacker TODO: 将Encoder改为Packer
	DefaultPacker string = "text" // encoding == text
	DefaultCodec  string = "json" // codec == text

	ErrorCode    string = "code"
	ErrorMessage string = "message"
	ErrorMore    string = "bin"
	ContextId    string = "context-id"
	CodecScheme  string = "codec"
	// PackerScheme TODO: 将Encoder改为Packer
	PackerScheme string = "packer"
)

const (
	_ScopeLength   = 2
	_ServiceName   = 1
	_Metadata      = 1
	_PayloadLayout = 1
)

type Getter interface {
	First() uint8
	BaseLength() int
	Length() uint32
	GetAndSetLength() uint32
	GetMsgType() uint8
	GetServiceName() string
	GetMsgId() uint64
	Payloads() container2.Slice[byte]
	PayloadsIterator() *container2.Iterator[[]byte]
	MinMux() int
}

type Setter interface {
	GetAndSetLength() uint32
	SetMsgType(msgTyp uint8)
	SetServiceName(serviceName string)
	SetMsgId(msgId uint64)
	AppendPayloads(p []byte)
	ReWritePayload(p []byte)
	SetPayloads(payloads []byte)
	Reset()
}

type GetterSetter interface {
	Getter
	Setter
}

func New() *Message {
	return &Message{
		MetaData:      container2.NewSliceMap[string, string](4),
		scope:         [2]uint8{MagicNumber},
		payloadLayout: make([]uint32, 0, 2),
		payloads:      nil,
	}
}

// Message 是对一次RPC调用传递的数据的描述
// 封装的方法均不是线程安全的
// DecodeHeader会修改一些内部值，调用时需要注意顺序
// 为了使用一致性的API，在访问内部一些简单的属性时，请使用Getxx方法
// 在设置一些值的过程中可能需要调整其它值，所以请使用Setxx方法
type Message struct {
	// int/uint数值统一使用大端序
	//	[0] == Magic (魔数，表示这是由littlerpc客户端发起或者服务端回复)
	//	[1] == MsgType (call/return & ping/pong)
	scope [2]uint8
	// 消息ID，用于表示一次完整的call/return的回复
	msgId uint64
	// 载荷数据的总长度
	payloadLength uint32
	// 实例名和调用方法名的布局
	//	Length (1 Byte) | ServiceName
	serviceName string
	// NOTE:
	//	有效载荷和元数据的范围
	//	在Mux模式中MetaData及其Desc必须能在一个MuxBlock下被装下，否则将会丢失数据
	// 元数据的布局
	//	NMetaData(1 Byte)|Key-Size(4 Byte)|Value-Size(4 Byte)|Key|Size
	// Example :
	//	"hello":"world","world:hello"
	// OutPut:
	//	0x00000002|0x00000005|0x00000005|hello|world|0x00000005|0x00000005|world|hello
	MetaData *container2.SliceMap[string, string]
	// 有效载荷数据的布局描述
	// Format :
	//	NArgs(1 Byte)|Arg1-Size(4 Byte)|Arg2-Size(4 Byte)|Arg3-Size(4 Byte)
	// Example :
	//	{"mypyload1":"haha"},{"mypyload2":"hehe"}
	// OutPut:
	//	0x00000002|0x00000014|0x00000014
	payloadLayout container2.Slice[uint32]
	// 调用参数序列化后的载荷数据
	// 如果被压缩了那么在反序列化时,最后剩下的数据均为参数载荷
	payloads container2.Slice[byte]
}

// BaseLength 获取基本数据的长度防止输入过短的数据导致panic
func (m *Message) BaseLength() int {
	ml := m.MinMux()
	// MinMux + NameLayout + NMetaData + NArgs
	return ml + _PayloadLayout + _ServiceName + _Metadata
}

// Length 根据结构计算序列化之后的数据长度, 长度存在时直接返回结果
// 不会设置m.payloadLength
func (m *Message) Length() uint32 {
	if m.payloadLength > 0 {
		return m.payloadLength
	}
	// desc : Scope & MsgId & PayloadLength & ServiceName & MetaData & Args
	baseLen := m.MinMux() + _ServiceName + _Metadata + _PayloadLayout
	// InstanceName & MethodName
	baseLen += len(m.serviceName)
	if m.MetaData != nil && m.MetaData.Len() > 0 {
		// Key&Value Struct MetaData
		baseLen += (m.MetaData.Len() * 4) * 2
		// Key & Value Size
		m.MetaData.Range(func(k, v string) bool {
			baseLen += len(k) + len(v)
			return true
		})
	}
	// NArgs
	baseLen += m.payloadLayout.Len() * 4
	baseLen += m.payloads.Len()
	return uint32(baseLen)
}

func (m *Message) GetAndSetLength() uint32 {
	m.payloadLength = m.Length()
	return m.Length()
}

func (m *Message) First() uint8 {
	return m.scope[0]
}

func (m *Message) GetMsgType() uint8 {
	return m.scope[1]
}

func (m *Message) SetMsgType(msgTyp uint8) {
	m.scope[1] = msgTyp
}

func (m *Message) GetServiceName() string {
	return m.serviceName
}

func (m *Message) SetServiceName(s string) {
	m.serviceName = s
}

func (m *Message) GetMsgId() uint64 {
	return m.msgId
}

func (m *Message) SetMsgId(msgId uint64) {
	m.msgId = msgId
}

func (m *Message) AppendPayloads(p []byte) {
	m.payloads = append(m.payloads, p...)
	m.payloadLayout = append(m.payloadLayout, uint32(len(p)))
}

func (m *Message) Payloads() container2.Slice[byte] {
	return m.payloads
}

func (m *Message) ReWritePayload(p []byte) {
	m.payloads.Reset()
	m.payloads.Append(p)
}

func (m *Message) SetPayloads(payloads []byte) {
	m.payloads = payloads
}

func (m *Message) PayloadsIterator() *container2.Iterator[[]byte] {
	rangCount := 0
	var length int
	if m.payloadLayout == nil || m.payloadLayout.Len() == 0 {
		length = 0
	} else {
		length = m.payloadLayout.Len()
	}
	return container2.NewIterator[[]byte](length, false, func(current int) []byte {
		if current+1 == m.payloadLayout.Len() {
			return m.payloads[rangCount : rangCount+int(m.payloadLayout[current])]
		}
		old := rangCount
		rangCount += int(m.payloadLayout[current])
		return m.payloads[old:rangCount]
	}, func() {
		rangCount = 0
	})
}

func (m *Message) MinMux() int {
	// Scope + MsgId + PayloadLength
	return len(m.scope) + 8 + 4
}

// Reset 给内存复用的操作提供一致性的语义
func (m *Message) Reset() {
	m.scope = [...]uint8{MagicNumber, 0}
	m.serviceName = ""
	m.msgId = 0
	m.payloadLength = 0
	m.MetaData.Reset()
	m.payloadLayout.Reset()
	m.payloads.Reset()
}

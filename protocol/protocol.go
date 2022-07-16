package protocol

import (
	"encoding/binary"
	"errors"
	"unsafe"
)

const (
	MagicNumber   uint8 = 0x45
	MessageCall   uint8 = 0x10
	MessageReturn uint8 = 0x18
	// MessageErrorReturn 调用因某种原因失败的返回
	MessageErrorReturn uint8 = 0x26
	// MessagePing Ping消息
	MessagePing uint8 = 0x33
	// MessagePong Pong消息
	MessagePong uint8 = 0x35

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

// MessageOperation 注意: 该接口主要作为操作Message的拓展功能
type MessageOperation interface {
	// UnmarshalHeader 从字节Slice中解码出header，并返回载荷数据的起始地址
	UnmarshalHeader(msg *Message, p []byte) (payloadStart int, err error)
	// RangePayloads 根据头提供的信息逐个遍历所有载荷数据
	// endAf指示是否是payloads中最后一个参数
	RangePayloads(msg *Message, p []byte, fn func(p []byte, endBefore bool) bool)
	// MarshalHeader 根据Msg Header编码出对应的字节Slice
	MarshalHeader(msg *Message, p *[]byte)
	// MarshalAll 序列化Header&Payloads
	MarshalAll(msg *Message, p *[]byte)
	// SetMetaData 设置对应的元数据
	SetMetaData(msg *Message, key, value string)
	// RangeMetaData 遍历所有元数据
	RangeMetaData(msg *Message, fn func(key string, value string))
	// Reset 指定策略的复用，对内存重用更加友好
	// resetOther指示是否释放|Scope|NameLayout|InstanceName|MethodName|MsgId|Timestamp
	// freeMetaData指示是否要释放存放元数据对应的map[string]sting
	// usePayload指示是否要复用载荷数据
	// useSize指示复用的slice类型长度的上限，即使指定了usePayload
	// payload数据超过这个长度还是会被释放
	Reset(msg *Message, resetOther, freeMetaData, usePayload bool, useSize int)
}

func NewMessageOperation() MessageOperation {
	return &messageOperationImpl{}
}

func NewMessage() *Message {
	return &Message{
		MetaData:      map[string]string{},
		PayloadLayout: make([]uint64, 0, 2),
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
	// 实例名和调用方法名的布局
	//	InstanceName-Size|MethodName-Size
	NameLayout [2]uint32
	// 实例名
	InstanceName string
	// 要调用的方法名
	MethodName string
	// 消息ID，用于跟踪等用途
	MsgId uint64
	// 生成该消息的时间戳,精确到毫秒
	Timestamp uint64
	// 有戏载荷和元数据的范围
	// 元数据的布局
	//	NMetaData(4 Byte)|Key-Size(4 Byte)|Value-Size(4 Byte)|Key|Size
	// Example :
	//	"hello":"world","world:hello"
	// OutPut:
	//	0x00000002|0x00000005|0x00000005|hello|world|0x00000005|0x00000005|world|hello
	MetaData map[string]string
	// 有效载荷数据的布局描述
	// Format :
	//	NArgs(4 Byte)|Arg1-Size(4 Byte)|Arg2-Size(4 Byte)|Arg3-Size(4 Byte)
	// Example :
	//	{"mypyload1":"haha"},{"mypyload2":"hehe"}
	// OutPut:
	//	0x00000002|0x00000014|0x00000014
	PayloadLayout []uint64
	Payloads      []byte
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

func (m *Message) GetTimestamp() uint64 {
	return m.Timestamp
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

func (m *Message) SetTimestamp(t uint64) {
	m.Timestamp = t
}

func (m *Message) AppendPayloads(p []byte) {
	m.Payloads = append(m.Payloads, p...)
	m.PayloadLayout = append(m.PayloadLayout, uint64(len(p)))
}

// Reset 给内存复用的操作提供一致性的语义
func (m *Message) Reset() {
	m.PayloadLayout = nil
	m.Payloads = nil
	*(*uint32)(unsafe.Pointer(&m.Scope)) = 0
	m.InstanceName = ""
	m.MethodName = ""
	m.MsgId = 0
	m.Timestamp = 0
	m.MetaData = nil
	m.PayloadLayout = nil
	m.Payloads = nil
}

type messageOperationImpl struct{}

func (m messageOperationImpl) UnmarshalHeader(msg *Message, p []byte) (payloadStart int, err error) {
	*(*uint32)(unsafe.Pointer(&msg.Scope)) = *(*uint32)(unsafe.Pointer(&p[0]))
	if msg.Scope[0] != MagicNumber {
		return -1, errors.New("not littlerpc protocol")
	}
	msg.NameLayout[0] = binary.BigEndian.Uint32(p[4:8])
	msg.NameLayout[1] = binary.BigEndian.Uint32(p[8:12])
	payloadStart += 12 + int(msg.NameLayout[0]) + int(msg.NameLayout[1])
	msg.InstanceName = string(p[12 : 12+msg.NameLayout[0]])
	msg.MethodName = string(p[12+msg.NameLayout[0] : payloadStart])
	msg.MsgId = binary.BigEndian.Uint64(p[payloadStart : payloadStart+8])
	msg.Timestamp = binary.BigEndian.Uint64(p[payloadStart+8 : payloadStart+16])
	payloadStart += 16
	// 有多少个元数据
	nMetaData := binary.BigEndian.Uint32(p[payloadStart:])
	payloadStart += 4
	for i := 0; i < int(nMetaData); i++ {
		keySize := binary.BigEndian.Uint32(p[payloadStart : payloadStart+4])
		valueSize := binary.BigEndian.Uint32(p[payloadStart+4 : payloadStart+8])
		if msg.MetaData == nil {
			msg.MetaData = make(map[string]string)
		}
		kss := payloadStart + 8
		vss := payloadStart + 8 + int(keySize)
		msg.MetaData[string(p[kss:vss])] = string(p[vss : vss+int(valueSize)])
		payloadStart += int(keySize+valueSize) + 8
	}
	nArgs := binary.BigEndian.Uint32(p[payloadStart:])
	// 为了保证更好的反序列化体验，如果不将layout置0的话
	// 会导致与Marshal/Unmarshal的结果重叠
	if msg.PayloadLayout != nil {
		msg.PayloadLayout = msg.PayloadLayout[:0]
	}
	payloadStart += 4
	for i := 0; i < int(nArgs); i++ {
		argsSize := binary.BigEndian.Uint64(p[payloadStart:])
		msg.PayloadLayout = append(msg.PayloadLayout, argsSize)
		payloadStart += 8
	}
	return payloadStart, nil
}

func (m messageOperationImpl) RangePayloads(msg *Message, p []byte, fn func(p []byte, endBefore bool) bool) {
	var i int
	nPayload := len(msg.PayloadLayout)
	for k, v := range msg.PayloadLayout {
		endAf := false
		if k == nPayload-1 {
			endAf = true
		}
		if !fn(p[i:i+int(v)], endAf) {
			return
		}
		i += int(v)
	}
}

func (m messageOperationImpl) MarshalHeader(msg *Message, p *[]byte) {
	*p = (*p)[:0]
	// 设置魔数值
	msg.Scope[0] = MagicNumber
	*p = append(*p, msg.Scope[:]...)
	*p = append(*p, FourBytesPadding...)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], msg.NameLayout[0])
	*p = append(*p, FourBytesPadding...)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], msg.NameLayout[1])
	*p = append(*p, msg.InstanceName...)
	*p = append(*p, msg.MethodName...)
	*p = append(*p, EightBytesPadding...)
	binary.BigEndian.PutUint64((*p)[len(*p)-8:], msg.MsgId)
	*p = append(*p, EightBytesPadding...)
	binary.BigEndian.PutUint64((*p)[len(*p)-8:], msg.Timestamp)
	// 序列化元数据
	*p = append(*p, FourBytesPadding...)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(len(msg.MetaData)))
	for k, v := range msg.MetaData {
		*p = append(*p, FourBytesPadding...)
		binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(len(k)))
		*p = append(*p, FourBytesPadding...)
		binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(len(v)))
		*p = append(*p, k...)
		*p = append(*p, v...)
	}
	// 序列化载荷数据描述信息
	*p = append(*p, FourBytesPadding...)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(len(msg.PayloadLayout)))
	for _, v := range msg.PayloadLayout {
		*p = append(*p, EightBytesPadding...)
		binary.BigEndian.PutUint64((*p)[len(*p)-8:], v)
	}
}

func (m messageOperationImpl) MarshalAll(msg *Message, p *[]byte) {
	m.MarshalHeader(msg, p)
	*p = append(*p, msg.Payloads...)
}

func (m messageOperationImpl) SetMetaData(msg *Message, key, value string) {
	if msg.MetaData == nil {
		msg.MetaData = make(map[string]string)
	}
	msg.MetaData[key] = value
}

func (m messageOperationImpl) RangeMetaData(msg *Message, fn func(key string, value string)) {
	if msg.MetaData == nil {
		return
	}
	for k, v := range msg.MetaData {
		fn(k, v)
	}
}

func (m messageOperationImpl) Reset(msg *Message, resetOther, freeMetaData, usePayload bool, useSize int) {
	if freeMetaData {
		msg.MetaData = nil
	}
	if len(msg.PayloadLayout) > useSize {
		msg.PayloadLayout = nil
	} else {
		msg.PayloadLayout = msg.PayloadLayout[:0]
	}
	if !usePayload {
		msg.Payloads = nil
	} else if usePayload && len(msg.Payloads) > useSize {
		msg.Payloads = nil
	} else {
		msg.Payloads = msg.Payloads[:0]
	}
	if resetOther {
		*(*uint32)(unsafe.Pointer(&msg.Scope)) = 0
		*(*uint64)(unsafe.Pointer(&msg.NameLayout)) = 0
		msg.MsgId = 0
		msg.Timestamp = 0
		msg.InstanceName = ""
		msg.MethodName = ""
	}
}

package message

import (
	"encoding/binary"
	"errors"
	"github.com/nyan233/littlerpc/pkg/container"
	. "github.com/nyan233/littlerpc/protocol"
	"math"
	"unsafe"
)

// MarshaToMux 此API只会序列化Mux功能需要的数据
func MarshaToMux(msg *Message, payloads *container.Slice[byte]) error {
	*payloads = (*payloads)[:msg.MinMux()]
	*payloads = append(*payloads, msg.scope[:]...)
	binary.BigEndian.PutUint64((*payloads)[4:12], msg.msgId)
	binary.BigEndian.PutUint32((*payloads)[12:16], msg.payloadLength)
	return nil
}

// UnmarshalFromMux 此API之后反序列化Mux功能所需要的数据
func UnmarshalFromMux(data container.Slice[byte], msg *Message) error {
	if data.Len() < msg.MinMux() {
		return errors.New("mux message is bad")
	}
	copy(msg.scope[:], data[:4])
	msg.msgId = binary.BigEndian.Uint64(data[4:12])
	msg.payloadLength = binary.BigEndian.Uint32(data[12:16])
	return nil
}

// Unmarshal 从字节Slice中解码出Message，并返回载荷数据的起始地址
func Unmarshal(p container.Slice[byte], msg *Message) error {
	if p.Len() == 0 || msg == nil {
		return errors.New("data or message is nil")
	}
	if p.Len() < msg.BaseLength() {
		return errors.New("data length < baseLen")
	}
	*(*uint32)(unsafe.Pointer(&msg.scope)) = *(*uint32)(unsafe.Pointer(&p[0]))
	if msg.scope[0] != MagicNumber {
		return errors.New("not littlerpc protocol")
	}
	p = p[4:]
	msg.msgId = binary.BigEndian.Uint64(p[:8])
	msg.payloadLength = binary.BigEndian.Uint32(p[8:12])
	p = p[12:]
	// NameLayout
	instanceLen := binary.BigEndian.Uint32(p[:4])
	methodNameLen := binary.BigEndian.Uint32(p[4:8])
	p = p[8:]
	if p.Len() < int(instanceLen) {
		return ErrBadMessage
	}
	msg.SetInstanceName(string(p[:instanceLen]))
	p = p[instanceLen:]
	if p.Len() < int(methodNameLen) {
		return ErrBadMessage
	}
	msg.SetMethodName(string(p[:methodNameLen]))
	p = p[methodNameLen:]
	// 有多少个元数据
	// 在可变长数据之后, 需要校验
	if p.Len() < 4 {
		return ErrBadMessage
	}
	nMetaData := binary.BigEndian.Uint32(p[:4])
	p = p[4:]
	for i := 0; i < int(nMetaData); i++ {
		if p.Len() < 8 {
			return ErrBadMessage
		}
		keySize := binary.BigEndian.Uint32(p[:4])
		valueSize := binary.BigEndian.Uint32(p[4:8])
		p = p[8:]
		// 相加防止溢出, 所以需要检查溢出
		if p.Len() < int(keySize+valueSize) || keySize > math.MaxUint32-valueSize {
			return ErrBadMessage
		}
		msg.MetaData.Store(string(p[:keySize]), string(p[keySize:keySize+valueSize]))
		p = p[keySize+valueSize:]
	}
	// 在可变长数据之后, 需要校验
	if p.Len() < 4 {
		return ErrBadMessage
	}
	nArgs := binary.BigEndian.Uint32(p[:4])
	p = p[4:]
	// 为了保证更好的反序列化体验，如果不将layout置0的话
	// 会导致与Marshal/Unmarshal的结果重叠
	if msg.payloadLayout != nil {
		msg.payloadLayout.Reset()
	}
	for i := 0; i < int(nArgs); i++ {
		if p.Len() < 4 {
			return ErrBadMessage
		}
		argsSize := binary.BigEndian.Uint32(p[:4])
		p = p[4:]
		msg.payloadLayout = append(msg.payloadLayout, argsSize)
	}
	// 不根据参数布局计算所有参数的载荷数据长度, 因为参数载荷数据可能会被压缩
	// 导致了长度不一致的情况
	msg.payloads.Reset()
	// 剩余的数据是载荷数据
	msg.payloads.Append(p)
	return nil
}

// Marshal 根据Msg Header编码出对应的字节Slice
// *[]byte是为了提供更好的内存复用语义
func Marshal(msg *Message, p *container.Slice[byte]) {
	p.Reset()
	msg.payloadLength = uint32(msg.Length())
	// 设置魔数值
	msg.scope[0] = MagicNumber
	*p = append(*p, msg.scope[:]...)
	p.Append(EightBytesPadding)
	binary.BigEndian.PutUint64((*p)[len(*p)-8:], msg.msgId)
	p.Append(FourBytesPadding)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], msg.payloadLength)
	p.Append(FourBytesPadding)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(len(msg.GetInstanceName())))
	p.Append(FourBytesPadding)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(len(msg.GetMethodName())))
	*p = append(*p, msg.GetInstanceName()...)
	*p = append(*p, msg.GetMethodName()...)
	// 序列化元数据
	p.Append(FourBytesPadding)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(msg.MetaData.Len()))
	msg.MetaData.Range(func(k, v string) bool {
		*p = append(*p, FourBytesPadding...)
		binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(len(k)))
		*p = append(*p, FourBytesPadding...)
		binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(len(v)))
		*p = append(*p, k...)
		*p = append(*p, v...)
		return true
	})
	// 序列化载荷数据描述信息
	p.Append(FourBytesPadding)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(len(msg.payloadLayout)))
	for _, v := range msg.payloadLayout {
		p.Append(FourBytesPadding)
		binary.BigEndian.PutUint32((*p)[len(*p)-4:], v)
	}
	p.Append(msg.payloads)
}

// ResetMsg 指定策略的复用，对内存重用更加友好
// resetOther指示是否释放|Scope|NameLayout|InstanceName|MethodName|MsgId|Timestamp
// freeMetaData指示是否要释放存放元数据对应的map[string]sting
// usePayload指示是否要复用载荷数据
// useSize指示复用的slice类型长度的上限，即使指定了usePayload
// payload数据超过这个长度还是会被释放
func ResetMsg(msg *Message, resetOther, freeMetaData, usePayload bool, useSize int) {
	if freeMetaData {
		msg.MetaData.Reset()
	}
	if len(msg.payloadLayout) > useSize {
		msg.payloadLayout = nil
	} else {
		msg.payloadLayout.Reset()
	}
	if !usePayload {
		msg.payloads = nil
	} else if usePayload && len(msg.payloads) > useSize {
		msg.payloads = nil
	} else {
		msg.payloads.Reset()
	}
	if resetOther {
		msg.scope = [...]uint8{MagicNumber, 0, 0, 0}
		msg.instanceName = ""
		msg.methodName = ""
		msg.msgId = 0
		msg.payloadLength = 0
	}
}

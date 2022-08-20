package protocol

import (
	"encoding/binary"
	"errors"
	"github.com/nyan233/littlerpc/container"
	"math"
	"unsafe"
)

/*
	Message操作的拓展方法
*/

var (
	ErrBadMessage = errors.New("bad message")
)

func MarshalMuxBlock(msg *MuxBlock, payloads *container.Slice[byte]) error {
	payloads.Reset()
	payloads.Append(make([]byte, MuxBlockBaseLen))
	(*payloads)[0] = msg.Flags
	binary.BigEndian.PutUint32((*payloads)[1:5], msg.StreamId)
	binary.BigEndian.PutUint64((*payloads)[5:13], msg.MsgId)
	binary.BigEndian.PutUint16((*payloads)[13:15], msg.PayloadLength)
	payloads.Append(msg.Payloads)
	return nil
}

func UnmarshalMuxBlock(data container.Slice[byte], msg *MuxBlock) error {
	if data.Len() < MuxBlockBaseLen {
		return ErrBadMessage
	}
	msg.Flags = data[0]
	data = data[1:]
	msg.StreamId = binary.BigEndian.Uint32(data[:4])
	data = data[4:]
	msg.MsgId = binary.BigEndian.Uint64(data[:8])
	data = data[8:]
	msg.PayloadLength = binary.BigEndian.Uint16(data[:2])
	msg.Payloads.Reset()
	msg.Payloads = data[2:]
	return nil
}

// MarshaMessageOnMux 此API只会序列化Mux功能需要的数据
func MarshaMessageOnMux(msg *Message, payloads *container.Slice[byte]) error {
	*payloads = (*payloads)[:msg.MinMux()]
	*payloads = append(*payloads, msg.Scope[:]...)
	binary.BigEndian.PutUint64((*payloads)[4:12], msg.MsgId)
	binary.BigEndian.PutUint32((*payloads)[12:16], msg.PayloadLength)
	return nil
}

// UnmarshalMessageOnMux 此API之后反序列化Mux功能所需要的数据
func UnmarshalMessageOnMux(data container.Slice[byte], msg *Message) error {
	if data.Len() < msg.MinMux() {
		return errors.New("mux message is bad")
	}
	copy(msg.Scope[:], data[:4])
	msg.MsgId = binary.BigEndian.Uint64(data[4:12])
	msg.PayloadLength = binary.BigEndian.Uint32(data[12:16])
	return nil
}

// UnmarshalMessage 从字节Slice中解码出Message，并返回载荷数据的起始地址
func UnmarshalMessage(p container.Slice[byte], msg *Message) error {
	if p.Len() == 0 || msg == nil {
		return errors.New("data or message is nil")
	}
	if p.Len() < msg.BaseLength() {
		return errors.New("data length < baseLen")
	}
	*(*uint32)(unsafe.Pointer(&msg.Scope)) = *(*uint32)(unsafe.Pointer(&p[0]))
	if msg.Scope[0] != MagicNumber {
		return errors.New("not littlerpc protocol")
	}
	p = p[4:]
	msg.MsgId = binary.BigEndian.Uint64(p[:8])
	msg.PayloadLength = binary.BigEndian.Uint32(p[8:12])
	p = p[12:]
	// NameLayout
	msg.NameLayout[0] = binary.BigEndian.Uint32(p[:4])
	msg.NameLayout[1] = binary.BigEndian.Uint32(p[4:8])
	p = p[8:]
	if p.Len() < int(msg.NameLayout[0]) {
		return ErrBadMessage
	}
	msg.InstanceName = string(p[:msg.NameLayout[0]])
	p = p[msg.NameLayout[0]:]
	if p.Len() < int(msg.NameLayout[1]) {
		return ErrBadMessage
	}
	msg.MethodName = string(p[:msg.NameLayout[1]])
	p = p[msg.NameLayout[1]:]
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
	if msg.PayloadLayout != nil {
		msg.PayloadLayout.Reset()
	}
	for i := 0; i < int(nArgs); i++ {
		if p.Len() < 4 {
			return ErrBadMessage
		}
		argsSize := binary.BigEndian.Uint32(p[:4])
		p = p[4:]
		msg.PayloadLayout = append(msg.PayloadLayout, argsSize)
	}
	// 不根据参数布局计算所有参数的载荷数据长度, 因为参数载荷数据可能会被压缩
	// 导致了长度不一致的情况
	msg.Payloads.Reset()
	// 剩余的数据是载荷数据
	msg.Payloads.Append(p)
	return nil
}

// RangePayloads 根据头提供的信息逐个遍历所有载荷数据
// endAf指示是否是payloads中最后一个参数
func RangePayloads(msg *Message, p container.Slice[byte], fn func(p []byte, endBefore bool) bool) {
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

// MarshalMessage 根据Msg Header编码出对应的字节Slice
// *[]byte是为了提供更好的内存复用语义
func MarshalMessage(msg *Message, p *container.Slice[byte]) {
	p.Reset()
	// 设置魔数值
	msg.Scope[0] = MagicNumber
	*p = append(*p, msg.Scope[:]...)
	p.Append(EightBytesPadding)
	binary.BigEndian.PutUint64((*p)[len(*p)-8:], msg.MsgId)
	p.Append(FourBytesPadding)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], msg.PayloadLength)
	p.Append(FourBytesPadding)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], msg.NameLayout[0])
	p.Append(FourBytesPadding)
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], msg.NameLayout[1])
	*p = append(*p, msg.InstanceName...)
	*p = append(*p, msg.MethodName...)
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
	binary.BigEndian.PutUint32((*p)[len(*p)-4:], uint32(len(msg.PayloadLayout)))
	for _, v := range msg.PayloadLayout {
		p.Append(FourBytesPadding)
		binary.BigEndian.PutUint32((*p)[len(*p)-4:], v)
	}
	p.Append(msg.Payloads)
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
	if len(msg.PayloadLayout) > useSize {
		msg.PayloadLayout = nil
	} else {
		msg.PayloadLayout.Reset()
	}
	if !usePayload {
		msg.Payloads = nil
	} else if usePayload && len(msg.Payloads) > useSize {
		msg.Payloads = nil
	} else {
		msg.Payloads.Reset()
	}
	if resetOther {
		*(*uint32)(unsafe.Pointer(&msg.Scope)) = 0
		*(*uint64)(unsafe.Pointer(&msg.NameLayout)) = 0
		msg.MsgId = 0
		msg.InstanceName = ""
		msg.MethodName = ""
	}
}

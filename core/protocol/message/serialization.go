package message

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"math"
)

// MarshaToMux 此API只会序列化Mux功能需要的数据
func MarshaToMux(msg *Message, payloads *container.Slice[byte]) error {
	if payloads.Cap() < msg.MinMux() {
		*payloads = make([]byte, 0, msg.MinMux())
	}
	msg.GetAndSetLength()
	*payloads = append(*payloads, msg.scope[:]...)
	*payloads = (*payloads)[:msg.MinMux()]
	binary.BigEndian.PutUint64((*payloads)[2:10], msg.msgId)
	binary.BigEndian.PutUint32((*payloads)[10:14], msg.payloadLength)
	return nil
}

// UnmarshalFromMux 此API之后反序列化Mux功能所需要的数据
func UnmarshalFromMux(data container.Slice[byte], msg *Message) error {
	if data.Len() < msg.MinMux() {
		return errors.New("mux message is bad")
	}
	copy(msg.scope[:], data[:_ScopeLength])
	msg.msgId = binary.BigEndian.Uint64(data[_ScopeLength:10])
	msg.payloadLength = binary.BigEndian.Uint32(data[10:14])
	return nil
}

// Unmarshal 从字节Slice中解码出Message，并返回载荷数据的起始地址
func Unmarshal(p container.Slice[byte], msg *Message) error {
	if p.Len() == 0 || msg == nil {
		return errors.New("data or message is nil")
	}
	if p.Len() < msg.BaseLength() {
		return errors.New("data Length < baseLen")
	}
	err := UnmarshalFromMux(p, msg)
	if err != nil {
		return err
	}
	p = p[msg.MinMux():]
	// NameLayout
	serviceNameLen := p[0]
	p = p[_ServiceName:]
	if p.Len() < int(serviceNameLen) {
		return errors.New("service name Length greater than p")
	}
	msg.SetServiceName(string(p[:serviceNameLen]))
	p = p[serviceNameLen:]
	// 有多少个元数据
	// 在可变长数据之后, 需要校验
	if p.Len() < _Metadata {
		return errors.New("p Length less than 1")
	}
	nMetaData := p[0]
	p = p[_Metadata:]
	for i := 0; i < int(nMetaData); i++ {
		if p.Len() < 8 {
			return errors.New("p Length less than 8 on nMetaData")
		}
		keySize := binary.BigEndian.Uint32(p[:4])
		valueSize := binary.BigEndian.Uint32(p[4:8])
		// 相加防止溢出, 所以需要检查溢出
		if p.Len() < int(keySize+valueSize) || keySize > math.MaxUint32-valueSize {
			return errors.New("key and value size overflow")
		}
		p = p[8:]
		msg.MetaData.Store(string(p[:keySize]), string(p[keySize:keySize+valueSize]))
		p = p[keySize+valueSize:]
	}
	// 在可变长数据之后, 需要校验
	if p.Len() < _PayloadLayout {
		return errors.New("p Length less than 4 on nArgs")
	}
	nArgs := p[0]
	p = p[_PayloadLayout:]
	// 为了保证更好的反序列化体验，如果不将layout置0的话
	// 会导致与Marshal/Unmarshal的结果重叠
	if msg.payloadLayout != nil {
		msg.payloadLayout.Reset()
	}
	for i := 0; i < int(nArgs); i++ {
		if p.Len() < 4 {
			return errors.New("p Length less than 4 on argument layout")
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
func Marshal(msg *Message, p *container.Slice[byte]) error {
	if err := MarshaToMux(msg, p); err != nil {
		return err
	}
	integerBuffer := make([]byte, 4)
	if len(msg.serviceName) > math.MaxUint8 {
		return errors.New(fmt.Sprintf("serviceName max Length = 255, but now Length = %d", len(msg.serviceName)))
	}
	p.AppendS(byte(len(msg.serviceName)))
	*p = append(*p, msg.serviceName...)
	// 序列化元数据
	if msg.MetaData.Len() > math.MaxUint8 {
		return errors.New(fmt.Sprintf("metaData max Length = 255, but now Length = %d", msg.MetaData.Len()))
	}
	p.AppendS(byte(msg.MetaData.Len()))
	msg.MetaData.Range(func(k, v string) bool {
		binary.BigEndian.PutUint32(integerBuffer, uint32(len(k)))
		p.Append(integerBuffer)
		binary.BigEndian.PutUint32(integerBuffer, uint32(len(v)))
		p.Append(integerBuffer)
		p.Append(convert.StringToBytes(k))
		p.Append(convert.StringToBytes(v))
		return true
	})
	// 序列化载荷数据描述信息
	if msg.payloadLayout.Len() > math.MaxUint8 {
		return errors.New(fmt.Sprintf("payloadLayout max Length = 255, but now Length = %d", msg.payloadLayout.Len()))
	}
	p.AppendS(byte(msg.payloadLayout.Len()))
	for _, v := range msg.payloadLayout {
		binary.BigEndian.PutUint32(integerBuffer, v)
		p.Append(integerBuffer)
	}
	p.Append(msg.payloads)
	return nil
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
		msg.scope = [...]uint8{MagicNumber, 0}
		msg.serviceName = ""
		msg.msgId = 0
		msg.payloadLength = 0
	}
}

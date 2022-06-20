package protocol

import (
	"bytes"
	"errors"
	"reflect"
	"strconv"
)

var (
	MessageHeaderNField = reflect.TypeOf(Header{}).NumField()
)

// Message 是对一次RPC调用传递的数据的描述
// 封装的方法均不是线程安全的
// DecodeHeader会修改一些内部值，调用时需要注意顺序
type Message struct {
	Header Header
	bodyIndex int
	BodyStart int
	Body      [][]byte
}


func (m *Message) DecodeHeader(data []byte) error {
	header := &m.Header
	headerBytes := bytes.SplitN(data,[]byte{';'},MessageHeaderNField)
	if len(headerBytes) != MessageHeaderNField {
		return errors.New("header format error")
	}
	header.MsgType = string(headerBytes[0])
	header.Encoding = string(headerBytes[1])
	header.CodecType = string(headerBytes[2])
	header.MethodName = string(headerBytes[3])
	msgId, err := strconv.ParseInt(string(headerBytes[4]),10,64)
	if err != nil {
		return err
	}
	timeStamp,err := strconv.ParseInt(string(headerBytes[5]), 10,64)
	if err != nil {
		return err
	}
	nBodyOffset,err := strconv.ParseInt(string(headerBytes[6]), 10,64)
	if err != nil {
		return err
	}
	header.MsgId = msgId
	header.Timestamp = timeStamp
	header.NBodyOffset = nBodyOffset
	var headerLen int
	for i := 0; i < MessageHeaderNField - 1;i++ {
		headerLen += len(headerBytes[i])
	}
	m.BodyStart = headerLen + (MessageHeaderNField - 1)
	// get body all offset
	bodyAllOffset := bytes.SplitN(headerBytes[7],[]byte{';'}, int(header.NBodyOffset) + 1)
	for _,v := range bodyAllOffset[:len(bodyAllOffset) - 1] {
		m.BodyStart += len(v) + 1
		offset,err := strconv.ParseInt(string(v),10,64)
		if err != nil {
			return err
		}
		header.BodyAllOffset = append(header.BodyAllOffset,int(offset))
	}
	return nil
}

func (m *Message) EncodeHeader() []byte {
	header := &m.Header
	buffer := make([]byte,0,128)
	headerTmp := make([][]byte,0,16)
	headerTmp = append(headerTmp,[]byte(header.MsgType))
	headerTmp = append(headerTmp,[]byte(header.Encoding))
	headerTmp = append(headerTmp,[]byte(header.CodecType))
	headerTmp = append(headerTmp,[]byte(header.MethodName))
	headerTmp = append(headerTmp,[]byte(strconv.FormatInt(header.MsgId,10)))
	headerTmp = append(headerTmp,[]byte(strconv.FormatInt(header.Timestamp,10)))
	headerTmp = append(headerTmp,[]byte(strconv.FormatInt(header.NBodyOffset,10)))
	for _,v := range headerTmp {
		buffer = append(buffer,v...)
		buffer = append(buffer,';')
	}
	for _,v := range m.Body {
		header.BodyAllOffset = append(header.BodyAllOffset,len(v))
	}
	for _,v := range header.BodyAllOffset {
		buffer = append(buffer,strconv.FormatInt(int64(v),10)...)
		buffer = append(buffer,';')
	}
	return buffer
}

func (m *Message) DecodeBodyFromBytes(data []byte) {
	data = data[m.BodyStart:]
	var offset int
	for _,v := range m.Header.BodyAllOffset {
		m.Body = append(m.Body,data[:offset+v])
	}
}

func (m *Message) DecodeBodyFromBodyBytes(data []byte) {
	var offset int
	for _,v := range m.Header.BodyAllOffset {
		m.Body = append(m.Body,data[offset:offset+v])
		offset += v
	}
}

func (m *Message) Encode(codec Codec,i interface{}) error {
	marshal, err := codec.Marshal(i)
	if err != nil {
		return err
	}
	m.Header.NBodyOffset++
	m.Body = append(m.Body,marshal)
	return nil
}

func (m *Message) EncodeRaw(p []byte) {
	m.Header.NBodyOffset++
	m.Body = append(m.Body,p)
}

func (m *Message) Decode(codec Codec,i interface{}) error {
	m.bodyIndex++
	return codec.Unmarshal(m.Body[m.bodyIndex - 1],i)
}

func (m *Message) ResetBody() {
	m.bodyIndex = 0
	m.Body = nil
}

func (m *Message) ResetHeader() {
	m.Header = Header{}
}

func (m *Message) ResetAll() {
	m.ResetHeader()
	m.ResetBody()
}
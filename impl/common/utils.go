package common

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/impl/transport"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/protocol"
	"runtime"
	"strconv"
	"time"
)

func HandleError(codec protocol.Codec,encoder packet.Encoder,msgId uint64, errNo protocol.Error, conn transport.ServerConnAdapter, appendInfo string) {
	md := protocol.FrameMd{
		ArgType:    protocol.Struct,
		AppendType: protocol.ServerError,
		Data:        nil,
	}
	sp := protocol.Body{}
	// write header
	header := protocol.Header{}
	header.Timestamp = uint64(time.Now().Unix())
	header.MsgType = protocol.MessageReturn
	header.MsgId = msgId
	header.CodecType = codec.Scheme()
	header.Encoding = encoder.Scheme()
	conn.Write(WriteHeader(header))
	switch errNo.Info {
	case ErrJsonUnMarshal.Info, ErrMethodNoRegister.Info, ErrCallArgsType.Info:
		errNo.Trace += appendInfo
		err := md.Encode(codec,errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Frame = append(sp.Frame, md)
		errNoBytes, err := codec.Marshal(&sp)
		if err != nil {
			panic(errors.New(fmt.Sprintf("codec/%s marshal failed",codec.Scheme())))
		}
		errNoBytes, err = encoder.EnPacket(errNoBytes)
		if err != nil {
			panic(errors.New(fmt.Sprintf("encoding/%s enpacket failed",encoder.Scheme())))
		}
		conn.Write(errNoBytes)
		break
	case ErrServer.Info:
		errNo.Info += appendInfo
		_, file, line, _ := runtime.Caller(1)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := md.Encode(codec,errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Frame = append(sp.Frame, md)
		errNoBytes, err := codec.Marshal(&sp)
		if err != nil {
			panic(errors.New(fmt.Sprintf("codec/%s marshal failed",codec.Scheme())))
		}
		errNoBytes, err = encoder.EnPacket(errNoBytes)
		if err != nil {
			panic(errors.New(fmt.Sprintf("encoding/%s enpacket failed",encoder.Scheme())))
		}
		conn.Write(errNoBytes)
	case Nil.Info:
		err := md.Encode(codec,errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Frame = append(sp.Frame, md)
		errNoBytes, err := codec.Marshal(&sp)
		if err != nil {
			panic(errors.New(fmt.Sprintf("codec/%s marshal failed",codec.Scheme())))
		}
		errNoBytes, err = encoder.EnPacket(errNoBytes)
		if err != nil {
			panic(errors.New(fmt.Sprintf("encoding/%s enpacket failed",encoder.Scheme())))
		}
		conn.Write(errNoBytes)
	}
}


func ReadHeader(data []byte) (protocol.Header,int) {
	header := &protocol.Header{}
	headerBytes := bytes.Split(data,[]byte{';'})
	header.MsgType = string(headerBytes[0])
	header.Encoding = string(headerBytes[1])
	header.CodecType = string(headerBytes[2])
	header.MethodName = string(headerBytes[3])
	msgId := [8]byte{}
	_, err := base64.StdEncoding.Decode(msgId[:], headerBytes[4])
	if err != nil {
		return protocol.Header{},0
	}
	header.MsgId = binary.BigEndian.Uint64(msgId[:])
	timestamp := [8]byte{}
	_, err = base64.StdEncoding.Decode(timestamp[:],headerBytes[5])
	if err != nil {
		return protocol.Header{},0
	}
	header.Timestamp = binary.BigEndian.Uint64(timestamp[:])
	var headerLen int
	for i := 0; i <= 5;i++ {
		headerLen += len(headerBytes[i])
	}
	return *header,headerLen + 6
}

func WriteHeader(header protocol.Header) []byte {
	buffer := make([]byte,0,128)
	headerTmp := make([][]byte,0,16)
	headerTmp = append(headerTmp,[]byte(header.MsgType))
	headerTmp = append(headerTmp,[]byte(header.Encoding))
	headerTmp = append(headerTmp,[]byte(header.CodecType))
	headerTmp = append(headerTmp,[]byte(header.MethodName))
	msgId := [8]byte{}
	binary.BigEndian.PutUint64(msgId[:],header.MsgId)
	msgIdBase64 := make([]byte,8 + 4)
	base64.StdEncoding.Encode(msgIdBase64,msgId[:])
	timestamp := [8]byte{}
	binary.BigEndian.PutUint64(timestamp[:],header.Timestamp)
	tsBase64 := make([]byte,8 + 4)
	base64.StdEncoding.Encode(tsBase64,timestamp[:])
	headerTmp = append(headerTmp,msgIdBase64)
	headerTmp = append(headerTmp,tsBase64)
	for _,v := range headerTmp {
		buffer = append(buffer,v...)
		buffer = append(buffer,';')
	}
	return buffer
}
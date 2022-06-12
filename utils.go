package littlerpc

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/nyan233/littlerpc/protocol"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

func HandleError(msg protocol.Message, errNo protocol.Error, conn *websocket.Conn, appendInfo string, more ...interface{}) {
	md := protocol.FrameMd{
		ArgType:    protocol.Struct,
		AppendType: protocol.ServerError,
		Data:        nil,
	}
	sp := msg.Body
	// write header
	header := msg.Header
	header.Timestamp = uint64(time.Now().Unix())
	header.MsgType = protocol.MessageReturn
	header.CodecType = protocol.DefaultCodecType
	header.Encoding = protocol.DefaultEncodingType
	conn.WriteMessage(websocket.TextMessage,writeHeader(header))
	switch errNo.Info {
	case ErrJsonUnMarshal.Info, ErrMethodNoRegister.Info, ErrCallArgsType.Info:
		errNo.Trace += appendInfo
		err := md.Encode(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Frame = append(sp.Frame, md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.WriteMessage(websocket.TextMessage, errNoBytes)
		break
	case ErrServer.Info:
		errNo.Info += appendInfo
		_, file, line, _ := runtime.Caller(1)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := md.Encode(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Frame = append(sp.Frame, md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.WriteMessage(websocket.TextMessage, errNoBytes)
	case Nil.Info:
		err := md.Encode(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Frame = append(sp.Frame, md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.WriteMessage(websocket.TextMessage, errNoBytes)
	}
}

// Little-RPC中定义了传入类型中不能为指针类型，所以Server根据这种方法就能快速判断
// 序列化好的远程栈帧的每个帧的类型是否和需要调用的参数列表的每个参数的类型相同
// 如果inputS有receiver的话，需要调用者对slice做Offset，比如[1:]
func checkInputTypeList(callArgs []reflect.Value, inputS []interface{}) (bool, []string) {
	if len(callArgs) != len(inputS) {
		return false, nil
	}
	for k := range callArgs {
		if !(callArgs[k].Type().Kind() == reflect.TypeOf(inputS[k]).Kind()) {
			return false, []string{callArgs[k].Type().Kind().String(),
				reflect.TypeOf(inputS[k]).Kind().String()}
		}
	}
	return true, nil
}

func readHeader(data []byte) (protocol.Header,int) {
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

func writeHeader(header protocol.Header) []byte {
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
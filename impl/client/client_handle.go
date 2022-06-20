package client

import (
	"github.com/nyan233/littlerpc/impl/internal"
	"github.com/nyan233/littlerpc/protocol"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

func (c *Client) handleProcessRetErr(msg *protocol.Message, i interface{}) (interface{}, error) {
	//单独处理返回的错误类型
	errBytes := msg.Body[len(msg.Body)-1]
	//处理最后返回的Error
	i, _ = lreflect.ToTypePtr(i)
	err := c.codec.UnmarshalError(errBytes, i)
	if err != nil {
		return nil, err
	}
	if e, ok := i.(*error); ok {
		return *e, nil
	}
	return i, nil
}

func (c *Client) readMsgAndDecodeReply(msg *protocol.Message, method reflect.Value, rep *[]interface{}) error {
	// 接收服务器返回的调用结果并将header反序列化
	buffer, err := c.conn.RecvData()
	// read header
	msg.ResetAll()
	err = msg.DecodeHeader(buffer)
	if err != nil {
		return err
	}
	// TODO : Client Handle Ping&Pong
	buffer, err = c.encoder.UnPacket(buffer[msg.BodyStart:])
	if err != nil {
		return err
	}
	// response body 不encoding/json来反序列化
	msg.DecodeBodyFromBodyBytes(buffer)
	if err != nil {
		return err
	}
	// 处理服务端传回的参数
	outputTypeList := lreflect.FuncOutputTypeList(method, false)
	for k, v := range msg.Body[:len(msg.Body)-1] {
		eface := outputTypeList[k]
		returnV, err := internal.CheckCoderType(c.codec, v, eface)
		if err != nil {
			return err
		}
		*rep = append(*rep, returnV)
	}
	// 处理返回值列表中最后的error
	returnV, err := c.handleProcessRetErr(msg, outputTypeList[len(outputTypeList)-1])
	if err != nil {
		return err
	}
	*rep = append(*rep, returnV)
	return nil
}

// return method
func (c *Client) identArgAndEncode(processName string, msg *protocol.Message, args []interface{}) (reflect.Value, error) {
	msg.Header.MethodName = processName
	methodData := strings.SplitN(processName, ".", 2)
	if len(methodData) != 2 || (methodData[0] == "" || methodData[1] == "") {
		panic("the illegal type name and method name")
	}
	method, ok := c.elem.Methods[methodData[1]]
	if !ok {
		panic("the method no register or is private method")
	}
	for _, v := range args {
		argType := internal.CheckIType(v)
		// 参数为指针类型则找出Elem的类型
		if argType == protocol.Pointer {
			argType = internal.CheckIType(reflect.ValueOf(v).Elem().Interface())
			// 不支持多重指针的数据结构
			if argType == protocol.Pointer {
				panic("multiple pointer no support")
			}
		}
		err := msg.Encode(c.codec, v)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
	}
	return method, nil
}

func (c *Client) sendCallMsg(msg *protocol.Message) error {
	// init header
	msg.Header.MsgId = rand.Int63()
	msg.Header.MsgType = protocol.MessageCall
	msg.Header.Timestamp = time.Now().Unix()
	msg.Header.Encoding = c.encoder.Scheme()
	msg.Header.CodecType = c.codec.Scheme()
	// request body
	memBuffer := c.memPool.Get().(*[]byte)
	*memBuffer = (*memBuffer)[:0]
	defer c.memPool.Put(memBuffer)
	// write header
	*memBuffer = append(*memBuffer, msg.EncodeHeader()...)
	bodyStart := len(*memBuffer)
	for _, v := range msg.Body {
		*memBuffer = append(*memBuffer, v...)
	}
	bodyBytes, err := c.encoder.EnPacket((*memBuffer)[bodyStart:])
	if err != nil {
		return err
	}
	// write body
	*memBuffer = append((*memBuffer)[:bodyStart], bodyBytes...)
	// write data
	_, err = c.conn.SendData(*memBuffer)
	if err != nil {
		return err
	}
	return nil
}

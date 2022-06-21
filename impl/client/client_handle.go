package client

import (
	"errors"
	"github.com/nyan233/littlerpc/impl/internal"
	"github.com/nyan233/littlerpc/protocol"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

func (c *Client) handleProcessRetErr(errBytes []byte, i interface{}) (interface{}, error) {
	//单独处理返回的错误类型
	//处理最后返回的Error
	i, _ = lreflect.ToTypePtr(i)
	err := c.codecWp.Instance().UnmarshalError(errBytes, i)
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
	c.mop.Reset(msg,false,false,true,4096)
	payloadStart,err := c.mop.UnmarshalHeader(msg,buffer)
	if err != nil {
		return err
	}
	// 处理服务器的错误返回
	if msg.GetMsgType() == protocol.MessageErrorReturn {
		return errors.New(string(buffer[payloadStart:]))
	}
	msg.Payloads = buffer[payloadStart:]
	// TODO : Client Handle Ping&Pong
	// encoder类型为text时不需要额外的内存拷贝
	// 默认的encoder即text
	if msg.GetEncoderType() != protocol.DefaultEncodingType {
		buffer, err = c.encoderWp.Instance().UnPacket(msg.Payloads)
		if err != nil {
			return err
		}
		msg.Payloads = append(msg.Payloads[:0],buffer...)
	}
	// 处理服务端传回的参数
	outputTypeList := lreflect.FuncOutputTypeList(method, false)
	var i int
	c.mop.RangePayloads(msg,msg.Payloads, func(p []byte,endBefore bool) bool {
		eface := outputTypeList[i]
		var returnV interface{}
		var err2 error
		if !endBefore {
			returnV, err2 = internal.CheckCoderType(c.codecWp.Instance(), p, eface)
			if err2 != nil {
				err = err2
				return false
			}
		} else {
			// 处理返回值列表中最后的error
			returnV, err2 = c.handleProcessRetErr(p, outputTypeList[len(outputTypeList)-1])
			if err2 != nil {
				err = err2
				return false
			}
		}
		*rep = append(*rep, returnV)
		i++
		return true
	})
	if err != nil {
		return err
	}
	return nil
}

// return method
func (c *Client) identArgAndEncode(processName string, msg *protocol.Message, args []interface{}) (reflect.Value, error) {
	methodData := strings.SplitN(processName, ".", 2)
	if len(methodData) != 2 || (methodData[0] == "" || methodData[1] == "") {
		panic("the illegal type name and method name")
	}
	msg.SetInstanceName(methodData[0])
	msg.SetMethodName(methodData[1])
	method, ok := c.elem.Methods[msg.MethodName]
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
		bytes,err := c.codecWp.Instance().Marshal(v)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		msg.AppendPayloads(bytes)
	}
	return method, nil
}

func (c *Client) sendCallMsg(msg *protocol.Message) error {
	// init header
	msg.SetMsgId(rand.Uint64())
	msg.SetMsgType(protocol.MessageCall)
	msg.SetTimestamp(uint64(time.Now().Unix()))
	msg.SetCodecType(uint8(c.codecWp.Index()))
	msg.SetEncoderType(uint8(c.encoderWp.Index()))
	// request body
	memBuffer := c.memPool.Get().(*[]byte)
	*memBuffer = (*memBuffer)[:0]
	defer c.memPool.Put(memBuffer)
	// write header
	// encoder类型为text不需要额外拷贝内存
	if c.encoderWp.Index() != int(protocol.DefaultEncodingType) {
		bodyBytes, err := c.encoderWp.Instance().EnPacket(msg.Payloads)
		if err != nil {
			return err
		}
		msg.Payloads = append(msg.Payloads[:0],bodyBytes...)
	}
	c.mop.MarshalAll(msg,memBuffer)
	// write data
	_, err := c.conn.SendData(*memBuffer)
	if err != nil {
		return err
	}
	return nil
}

package client

import (
	"errors"
	"github.com/nyan233/littlerpc/impl/common"
	"github.com/nyan233/littlerpc/impl/internal"
	"github.com/nyan233/littlerpc/impl/transport"
	"github.com/nyan233/littlerpc/middle/balance"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/middle/resolver"
	"github.com/nyan233/littlerpc/protocol"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"github.com/zbh255/bilog"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	addrCollection []string
	mu sync.Mutex
)

type Client struct {
	mu sync.Mutex
	elem   common.ElemMeta
	logger bilog.Logger
	// 错误处理回调函数
	onErr func(err error)
	// client Engine
	conn transport.ClientTransport
	// 简单的内存池
	memPool sync.Pool
	// 字节流编码器
	encoder packet.Encoder
	// 结构化数据编码器
	codec protocol.Codec
}

func ClientOpenBalance(scheme,url string,updateT time.Duration) error {
	mu.Lock()
	defer mu.Unlock()
	rb := resolver.GetResolver(scheme)
	if rb == nil {
		return errors.New("no this resolver scheme")
	}
	rb.SetOpen(true)
	rb.SetUpdateTime(updateT)
	addrC,err := rb.Instance().Parse(url)
	if err != nil {
		return err
	}
	addrCollection = addrC
	return nil
}

func NewClient(opts ...clientOption) (*Client,error) {
	config := &Config{}
	WithDefaultClient()(config)
	for _, v := range opts {
		v(config)
	}
	client := &Client{}
	client.logger = config.Logger
	if config.CallOnErr != nil {
		client.onErr = config.CallOnErr
	} else {
		client.onErr = client.defaultOnErr
	}
	// TODO 配置解析器和负载均衡器
	if config.BalanceScheme != "" {
		mu.Lock()
		balancer := balance.GetBalancer(config.BalanceScheme)
		if balancer == nil {
			panic("no balancer scheme")
		}
		addr := balancer.Target(addrCollection)
		mu.Unlock()
		config.ServerAddr = addr
		conn,err := clientSupportCollection[config.NetWork](*config)
		if err != nil {
			return nil,err
		}
		client.conn = conn
	} else {
		conn,err := clientSupportCollection[config.NetWork](*config)
		if err != nil {
			return nil,err
		}
		client.conn = conn
	}
	// init pool
	client.memPool = sync.Pool{
		New: func() interface{} {
			tmp := make([]byte,4096)
			return &tmp
		},
	}
	// encoder
	client.encoder = config.Encoder
	// codec
	client.codec = config.Codec
	return client,nil
}

func (c *Client) BindFunc(i interface{}) error {
	if i == nil {
		return errors.New("register elem is nil")
	}
	// init message id in rand
	rand.Seed(time.Now().UnixNano())
	elemD := common.ElemMeta{}
	elemD.Typ = reflect.TypeOf(i)
	elemD.Data = reflect.ValueOf(i)
	// init map
	elemD.Methods = make(map[string]reflect.Value, elemD.Typ.NumMethod())
	for i := 0; i < elemD.Typ.NumMethod(); i++ {
		method := elemD.Typ.Method(i)
		if method.IsExported() {
			elemD.Methods[method.Name] = method.Func
		}
	}
	c.elem = elemD
	return nil
}

func (c *Client) defaultOnErr(err error) {
	c.logger.ErrorFromErr(err)
}

func (c *Client) Call(processName string, args ...interface{}) (rep []interface{}, uErr error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	methodData := strings.SplitN(processName,".",2)
	if len(methodData) != 2 || (methodData[0] == "" || methodData[1] == "") {
		panic("the illegal type name and method name")
	}
	msg := &protocol.Message{}
	msg.Header.MethodName = processName
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

		err := msg.Encode(c.codec,v)
		if err != nil {
			panic(err)
		}
	}
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
	*memBuffer = append(*memBuffer,msg.EncodeHeader()...)
	bodyStart := len(*memBuffer)
	for _,v := range msg.Body {
		*memBuffer = append(*memBuffer,v...)
	}
	bodyBytes, err := c.encoder.EnPacket((*memBuffer)[bodyStart:])
	if err != nil {
		c.onErr(err)
		return
	}
	// write body
	*memBuffer = append((*memBuffer)[:bodyStart],bodyBytes...)
	// write data
	if c.encoder.Scheme() == "text" {
		_,err = c.conn.SendData(*memBuffer)
	} else {
		_,err = c.conn.SendData(*memBuffer)
	}
	if err != nil {
		c.onErr(err)
		return
	}
	// 接收服务器返回的调用结果并将header反序列化
	buffer, err := c.conn.RecvData()
	// read header
	msg.ResetAll()
	err = msg.DecodeHeader(buffer)
	if err != nil {
		c.onErr(err)
		return
	}
	// TODO : Client Handle Ping&Pong
	buffer,err = c.encoder.UnPacket(buffer[msg.BodyStart:])
	if err != nil {
		c.onErr(err)
		return
	}
	// response body 暂时需要encoding/json来反序列化
	msg.DecodeBodyFromBodyBytes(buffer)
	if err != nil {
		c.onErr(err)
		return
	}
	// 处理服务端传回的参数
	outputTypeList := lreflect.FuncOutputTypeList(method,false)
	for k, v := range msg.Body[:len(msg.Body)-1] {
		eface := outputTypeList[k]
		returnV, err := internal.CheckCoderType(c.codec,v, eface)
		if err != nil {
			c.onErr(err)
			return
		}
		rep = append(rep, returnV)
	}
	// 单独处理返回的错误类型
	//errMd := msg.Body[len(msg.Body)-1]
	// 处理最后返回的Error
	// 返回的数据的类型不可能是指针类型，需要客户端自己去处理
	//switch errMd.ArgType {
	//case protocol.Struct:
	//	if c.codec.Scheme() != "json" {
	//		break
	//	}
	//	errPtr := &protocol.Error{}
	//	ierr := c.codec.Unmarshal(errMd.Data, errPtr)
	//	if ierr != nil {
	//		panic(err)
	//	}
	//	uErr = errPtr
	//case protocol.String:
	//	if c.codec.Scheme() != "json" {
	//		break
	//	}
	//	var tmp = ""
	//	err := c.codec.Unmarshal(errMd.Data, &tmp)
	//	if err != nil {
	//		panic(err)
	//	}
	//	uErr = errors.New(tmp)
	//case protocol.Integer:
	//	if c.codec.Scheme() != "json" {
	//		break
	//	}
	//	uErr = error(nil)
	//}
	//// 根据Header CallType判断是否服务器错误
	//// 检查错误是Server的异常还是远程过程正常返回的error
	//if errMd.AppendType == protocol.ServerError {
	//	c.onErr(uErr)
	//	uErr = nil
	//}
	return
}

func (c *Client) Close() error {
	return c.conn.Close()
}

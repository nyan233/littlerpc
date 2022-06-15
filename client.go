package littlerpc

import (
	"encoding/json"
	"errors"
	"github.com/nyan233/littlerpc/internal/transport"
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
	elem   ElemMeta
	logger bilog.Logger
	// 错误处理回调函数
	onErr func(err error)
	// client Engine
	conn *transport.WebSocketTransClient
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

func NewClient(opts ...clientOption) *Client {
	config := &ClientConfig{}
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
		conn := transport.NewWebSocketTransClient(config.TlsConfig, addr)
		client.conn = conn
	} else {
		conn := transport.NewWebSocketTransClient(config.TlsConfig, config.ServerAddr)
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
	return client
}

func (c *Client) BindFunc(i interface{}) error {
	if i == nil {
		return errors.New("register elem is nil")
	}
	// init message id in rand
	rand.Seed(time.Now().UnixNano())
	elemD := ElemMeta{}
	elemD.typ = reflect.TypeOf(i)
	elemD.data = reflect.ValueOf(i)
	// init map
	elemD.methods = make(map[string]reflect.Value, elemD.typ.NumMethod())
	for i := 0; i < elemD.typ.NumMethod(); i++ {
		method := elemD.typ.Method(i)
		if method.IsExported() {
			elemD.methods[method.Name] = method.Func
		}
	}
	c.elem = elemD
	return nil
}

func (c *Client) defaultOnErr(err error) {
	c.logger.ErrorFromErr(err)
}

func (c *Client) Call(processName string, args ...interface{}) (rep []interface{}, uErr error) {
	methodData := strings.SplitN(processName,".",2)
	if len(methodData) != 2 || (methodData[0] == "" || methodData[1] == "") {
		panic("the illegal type name and method name")
	}
	msg := &protocol.Message{}
	msg.Header.MethodName = processName
	method, ok := c.elem.methods[methodData[1]]
	if !ok {
		panic("the method no register or is private method")
	}
	for _, v := range args {
		var md protocol.FrameMd
		md.ArgType = checkIType(v)
		// 参数为指针类型则找出Elem的类型
		if md.ArgType == protocol.Pointer {
			md.ArgType = checkIType(reflect.ValueOf(v).Elem().Interface())
			// 不支持多重指针的数据结构
			if md.ArgType == protocol.Pointer {
				panic("multiple pointer no support")
			}
		}
		// 参数为数组类型则保证额外的类型
		if md.ArgType == protocol.Array {
			md.AppendType = checkIType(lreflect.IdentArrayOrSliceType(v))
		}

		argBytes, err := c.codec.Marshal(v)
		if err != nil {
			panic(err)
		}
		md.Data = argBytes
		msg.Body.Frame = append(msg.Body.Frame, md)
	}
	// init header
	msg.Header.MsgId = rand.Uint64()
	msg.Header.MsgType = protocol.MessageCall
	msg.Header.Timestamp = uint64(time.Now().Unix())
	msg.Header.Encoding = c.encoder.Scheme()
	msg.Header.CodecType = c.codec.Scheme()
	// request body 暂时需要encoding/json来序列化
	requestBytes, err := json.Marshal(msg.Body)
	if err != nil {
		panic(err)
	}
	memBuffer := c.memPool.Get().(*[]byte)
	*memBuffer = (*memBuffer)[:0]
	defer c.memPool.Put(memBuffer)
	// write header
	*memBuffer = append(*memBuffer,writeHeader(msg.Header)...)
	requestBytes, err = c.encoder.EnPacket(requestBytes)
	if err != nil {
		c.onErr(err)
		return
	}
	// write body
	*memBuffer = append(*memBuffer,requestBytes...)
	// write data
	if c.encoder.Scheme() == "text" {
		err = c.conn.WriteTextMessage(*memBuffer)
	} else {
		err = c.conn.WriteBinaryMessage(*memBuffer)
	}
	if err != nil {
		c.onErr(err)
		return
	}
	// 接收服务器返回的调用结果并将header反序列化
	_, buffer, err := c.conn.RecvMessage()
	// read header
	header,headerLen := readHeader(buffer)
	// TODO : Client Handle Ping&Pong
	_ = header
	msg.Body.Frame = nil
	buffer,err = c.encoder.UnPacket(buffer[headerLen:])
	if err != nil {
		c.onErr(err)
		return
	}
	// response body 暂时需要encoding/json来反序列化
	err = json.Unmarshal(buffer, &msg.Body)
	if err != nil {
		c.onErr(err)
		return
	}
	// 处理服务端传回的参数
	outputTypeList := lreflect.FuncOutputTypeList(method)
	for k, v := range msg.Body.Frame[:len(msg.Body.Frame)-1] {
		eface := outputTypeList[k]
		md := protocol.FrameMd{
			ArgType:    v.ArgType,
			AppendType: v.AppendType,
			Data:        v.Data,
		}
		returnV, err := checkCoderType(c.codec,md, eface)
		if err != nil {
			c.onErr(err)
			return
		}
		rep = append(rep, returnV)
	}
	// 单独处理返回的错误类型
	errMd := msg.Body.Frame[len(msg.Body.Frame)-1]
	// 处理最后返回的Error
	// 返回的数据的类型不可能是指针类型，需要客户端自己去处理
	switch errMd.ArgType {
	case protocol.Struct:
		if c.codec.Scheme() != "json" {
			break
		}
		errPtr := &protocol.Error{}
		ierr := c.codec.Unmarshal(errMd.Data, errPtr)
		if ierr != nil {
			panic(err)
		}
		uErr = errPtr
	case protocol.String:
		if c.codec.Scheme() != "json" {
			break
		}
		var tmp = ""
		err := c.codec.Unmarshal(errMd.Data, &tmp)
		if err != nil {
			panic(err)
		}
		uErr = errors.New(tmp)
	case protocol.Integer:
		if c.codec.Scheme() != "json" {
			break
		}
		uErr = error(nil)
	}
	// 检查错误是Server的异常还是远程过程正常返回的error
	if errMd.AppendType == protocol.ServerError {
		c.onErr(uErr)
		uErr = nil
	}
	return
}

func (c *Client) Close() error {
	return c.conn.Close()
}

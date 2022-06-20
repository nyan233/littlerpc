package client

import (
	"errors"
	"github.com/nyan233/littlerpc/impl/common"
	"github.com/nyan233/littlerpc/impl/transport"
	"github.com/nyan233/littlerpc/middle/balance"
	"github.com/nyan233/littlerpc/middle/packet"
	"github.com/nyan233/littlerpc/middle/resolver"
	"github.com/nyan233/littlerpc/protocol"
	"github.com/zbh255/bilog"
	"math/rand"
	"reflect"
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
	// client Engine
	conn transport.ClientTransport
	// 简单的内存池
	memPool sync.Pool
	// 字节流编码器
	encoder packet.Encoder
	// 结构化数据编码器
	codec protocol.Codec
}

func OpenBalance(scheme,url string,updateT time.Duration) error {
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


// Call 远程过程返回的所有值都在rep中,sErr是调用过程中的错误，不是远程过程返回的错误
// 现在的onErr回调函数将不起作用，sErr表示Client.Call()在调用一些函数返回的错误或者调用远程过程时返回的错误
// 用户定义的远程过程返回的错误应该被安排在rep的最后一个槽位中
// 生成器应该将优先将sErr错误返回
func (c *Client) Call(processName string, args ...interface{}) (rep []interface{}, sErr error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	msg := &protocol.Message{}
	method,err := c.identArgAndEncode(processName,msg,args)
	if err != nil {
		return nil,err
	}
	err = c.sendCallMsg(msg)
	if err != nil {
		return nil, err
	}
	err = c.readMsgAndDecodeReply(msg,method,&rep)
	if err != nil {
		return nil,err
	}
	return
}

func (c *Client) Close() error {
	return c.conn.Close()
}

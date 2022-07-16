package client

import (
	"errors"
	"github.com/nyan233/littlerpc/common"
	"github.com/nyan233/littlerpc/common/transport"
	"github.com/nyan233/littlerpc/middle/resolver"
	"github.com/nyan233/littlerpc/protocol"
	"reflect"
	"sync"
	"time"
)

var (
	addrCollection []string
	mu             sync.Mutex
)

type lockConn struct {
	mu   *sync.Mutex
	conn transport.ClientTransport
}

func OpenBalance(scheme, url string, updateT time.Duration) error {
	mu.Lock()
	defer mu.Unlock()
	rb := resolver.GetResolver(scheme)
	if rb == nil {
		return errors.New("no this resolver scheme")
	}
	rb.SetOpen(true)
	rb.SetUpdateTime(updateT)
	addrC, err := rb.Instance().Parse(url)
	if err != nil {
		return err
	}
	addrCollection = addrC
	return nil
}

func (c *Client) BindFunc(instanceName string, i interface{}) error {
	if i == nil {
		return errors.New("register elem is nil")
	}
	if instanceName == "" {
		return errors.New("the typ name is not defined")
	}
	elemD := common.ElemMeta{}
	elemD.Typ = reflect.TypeOf(i)
	elemD.Data = reflect.ValueOf(i)
	// init map
	elemD.Methods = make(map[string]reflect.Value, elemD.Typ.NumMethod())
	// NOTE: 这里的判断不能依靠map的len/cap来确定实例用于多少的绑定方法
	// 因为len/cap都不能提供准确的信息,调用make()时指定的cap只是给真正创建map的函数一个提示
	// 并不代表真实大小，对没有插入过数据的map调用len()永远为0
	if elemD.Typ.NumMethod() == 0 {
		return errors.New("instance no method")
	}
	for i := 0; i < elemD.Typ.NumMethod(); i++ {
		method := elemD.Typ.Method(i)
		if method.IsExported() {
			elemD.Methods[method.Name] = method.Func
		}
	}
	c.elems.Store(instanceName, elemD)
	return nil
}

// Call 远程过程返回的所有值都在rep中,sErr是调用过程中的错误，不是远程过程返回的错误
// 现在的onErr回调函数将不起作用，sErr表示Client.Call()在调用一些函数返回的错误或者调用远程过程时返回的错误
// 用户定义的远程过程返回的错误应该被安排在rep的最后一个槽位中
// 生成器应该将优先将sErr错误返回
func (c *Client) Call(processName string, args ...interface{}) (rep []interface{}, sErr error) {
	conn := getConnFromMux(c)
	conn.mu.Lock()
	defer conn.mu.Unlock()
	msg := protocol.NewMessage()
	method, err := c.identArgAndEncode(processName, msg, args)
	if err != nil {
		return nil, err
	}
	err = c.sendCallMsg(msg, conn.conn)
	if err != nil {
		return nil, err
	}
	err = c.readMsgAndDecodeReply(msg, conn.conn, method, &rep)
	if err != nil {
		return nil, err
	}
	return
}

// AsyncCall 该函数返回时至少数据已经经过Codec的序列化，调用者有责任检查error
// 该函数可能会传递来自Codec和内部组件的错误，因为它在发送消息之前完成
func (c *Client) AsyncCall(processName string, args ...interface{}) error {
	msg := protocol.NewMessage()
	method, err := c.identArgAndEncode(processName, msg, args)
	if err != nil {
		return err
	}
	go func() {
		// 查找对应的回调函数
		var callBackIsOk bool
		cbFn, ok := c.callBacks.Load(processName)
		callBackIsOk = ok
		// 在池中获取一个底层传输的连接
		conn := getConnFromMux(c)
		conn.mu.Lock()
		defer mu.Unlock()
		err := c.sendCallMsg(msg, conn.conn)
		if err != nil && callBackIsOk {
			cbFn(nil, err)
			return
		} else if err != nil && !callBackIsOk {
			return
		}
		rep := make([]interface{}, 0)
		err = c.readMsgAndDecodeReply(msg, conn.conn, method, &rep)
		if err != nil && callBackIsOk {
			cbFn(nil, err)
			return
		} else if err != nil && !callBackIsOk {
			return
		}
		if callBackIsOk {
			cbFn(rep, nil)
		}
	}()
	return nil
}

func (c *Client) RegisterCallBack(processName string, fn func(rep []interface{}, err error)) {
	c.callBacks.Store(processName, fn)
}

func (c *Client) Close() error {
	c.sendGPool.Stop()
	for _, v := range c.concurrentConnect {
		v.mu.Lock()
		err := v.conn.Close()
		v.mu.Unlock()
		if err != nil {
			v.mu.Unlock()
			return err
		}
	}
	return nil
}

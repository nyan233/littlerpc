package context

import (
	"context"
	"github.com/nyan233/littlerpc/core/container"
	"net"
	"time"
)

// Context 非线程安全, 跨线程传递需要调用Clone
type Context struct {
	OriginCtx    context.Context
	Header       *container.SliceMap[string, string]
	RemoteAddr   net.Addr
	LocalAddr    net.Addr
	ServiceName  string
	localStorage map[any]any
}

func NewContext(ctx context.Context) *Context {
	return &Context{OriginCtx: ctx, Header: container.NewSliceMap[string, string](16), localStorage: make(map[any]any, 4)}
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.OriginCtx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.OriginCtx.Done()
}

func (c *Context) Err() error {
	return c.OriginCtx.Err()
}

func (c *Context) Value(key any) any {
	val, _ := c.localStorage[key]
	return val
}

func (c *Context) SetValue(key, value any) {
	c.localStorage[key] = value
}

func (c *Context) Clone() *Context {
	newCtx := NewContext(c.OriginCtx)
	newCtx.RemoteAddr = c.RemoteAddr
	newCtx.LocalAddr = c.LocalAddr
	c.Header.Range(func(k string, v string) (next bool) {
		newCtx.Header.Store(k, v)
		return true
	})
	newCtx.localStorage = make(map[any]any, len(c.localStorage))
	for key, val := range c.localStorage {
		newCtx.localStorage[key] = val
	}
	return newCtx
}

func Background() *Context {
	return NewContext(context.Background())
}

func TODO() *Context {
	return NewContext(context.TODO())
}

func WithCancelOfOrigin(ctx context.Context) (*Context, func()) {
	newOCtx, cancelFn := context.WithCancel(ctx)
	newCtx := NewContext(newOCtx)
	return newCtx, cancelFn
}

func WithCancel(ctx *Context) (*Context, func()) {
	newOCtx, cancelFn := context.WithCancel(ctx.OriginCtx)
	newCtx := ctx.Clone()
	newCtx.OriginCtx = newOCtx
	newCtx.localStorage = ctx.localStorage
	return newCtx, cancelFn
}

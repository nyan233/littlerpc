package context

import (
	"context"
	"net"
	"time"
)

type remoteAddr struct{}
type localAddr struct{}
type initData struct{}
type InitData struct {
	Start       time.Time
	ServiceName string
	MsgType     uint8
}

func WithRemoteAddr(ctx context.Context, addr net.Addr) context.Context {
	return context.WithValue(ctx, remoteAddr{}, addr)
}

func CheckRemoteAddr(ctx context.Context) net.Addr {
	a, _ := ctx.Value(remoteAddr{}).(net.Addr)
	return a
}

func WithLocalAddr(ctx context.Context, addr net.Addr) context.Context {
	return context.WithValue(ctx, localAddr{}, addr)
}

func CheckLocalAddr(ctx context.Context) net.Addr {
	a, _ := ctx.Value(localAddr{}).(net.Addr)
	return a
}

func WithInitData(ctx context.Context, p *InitData) context.Context {
	return context.WithValue(ctx, initData{}, p)
}

func CheckInitData(ctx context.Context) *InitData {
	a, _ := ctx.Value(initData{}).(*InitData)
	return a
}

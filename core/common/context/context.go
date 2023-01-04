package context

import (
	"context"
	"net"
)

type remoteAddr struct{}
type localAddr struct{}
type service struct{}

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

func WithService(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, service{}, s)
}

func CheckService(ctx context.Context) string {
	a, _ := ctx.Value(service{}).(string)
	return a
}

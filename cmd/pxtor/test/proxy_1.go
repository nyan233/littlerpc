package test

import (
	"github.com/nyan233/littlerpc/core/server"
)

type Test struct {
	Name string
	Key  string
	Uid  uint64
	server.RpcServer
}

func (p *Test) Setup() {
	err := p.HijackProcess("Foo", func(stub *server.Stub) {
		return
	})
	if err != nil {
		panic(err)
	}
}

func (p *Test) Foo(s1 string) (int, error) {
	return 1 << 20, nil
}

func (p *Test) Bar(s1 string) (int, error) {
	return 1 << 30, nil
}

func (p *Test) NoReturnValue(i int) error {
	return nil
}

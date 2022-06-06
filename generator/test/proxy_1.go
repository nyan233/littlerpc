package test

import (
	"errors"
	"github.com/nyan233/littlerpc/coder"
)

type Test struct {}

func (p *Test) Foo(s1 string) int {
	return 1 << 20
}

func (p *Test) Bar(s1 string) int {
	return 1 << 30
}

func (p *Test) NoReturnValue(i int) {}

func (p *Test) ErrHandler(s1 string) (err error) {
	return errors.New(s1)
}

func (p *Test) ErrHandler2(s1 string) (err *coder.Error) {
	return nil
}
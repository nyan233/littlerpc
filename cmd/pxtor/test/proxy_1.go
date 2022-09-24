package test

import (
	"errors"
	"github.com/nyan233/littlerpc/pkg/common"
	perror "github.com/nyan233/littlerpc/protocol/error"
)

type Test struct{}

func (p *Test) Foo(s1 string) (int, error) {
	return 1 << 20, nil
}

func (p *Test) Bar(s1 string) (int, error) {
	return 1 << 30, nil
}

func (p *Test) NoReturnValue(i int) error {
	return nil
}

func (p *Test) ErrHandler(s1 string) (err error) {
	return errors.New(s1)
}

func (p *Test) ErrHandler2(s1 string) (err error) {
	return perror.LNewStdError(common.Success.Code(), common.Success.Message())
}

package test

import (
	"errors"
	"github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/container"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
)

func (p *Test) ErrHandler(ctx *context.Context, s1 string) (err error) {
	return errors.New(s1)
}

func (p *Test) ErrHandler2(ctx *context.Context, s1 string) (err error) {
	return perror.LNewStdError(errorhandler.Success.Code(), errorhandler.Success.Message())
}

func (p *Test) ImportTest(ctx *context.Context, l1 container.ByteSlice) (err error) {
	return nil
}

func (p *Test) ImportTest2(ctx *context.Context, l1 container.ByteSlice, l2 *container.ByteSlice) (err error) {
	return nil
}

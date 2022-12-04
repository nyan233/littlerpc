package test

import (
	"errors"
	"github.com/nyan233/littlerpc/pkg/common/errorhandler"
	perror "github.com/nyan233/littlerpc/protocol/error"
)

func (p *Test) ErrHandler(s1 string) (err error) {
	return errors.New(s1)
}

func (p *Test) ErrHandler2(s1 string) (err error) {
	return perror.LNewStdError(errorhandler.Success.Code(), errorhandler.Success.Message())
}

package common

import (
	perror "github.com/nyan233/littlerpc/protocol/error"
)

var DefaultErrHandler = &JsonErrorHandler{}

type JsonErrorHandler struct{}

func (j JsonErrorHandler) LNewErrorDesc(code int, message string, mores ...interface{}) perror.LErrorDesc {
	return perror.LNewStdError(code, message, mores...)
}

func (j JsonErrorHandler) LWarpErrorDesc(desc perror.LErrorDesc, mores ...interface{}) perror.LErrorDesc {
	return perror.LWarpStdError(desc, mores...)
}

package auth

import (
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
)

const (
	NameKey     = "user"
	PasswordKey = "password"
)

var (
	ErrorFailed = errorhandler.DefaultErrHandler.LNewErrorDesc(10024, "user_name or password never correct")
)

type LRPCAuthorization struct {
	plugin.Abstract
	UserName, Password string
}

func NewBasicAuth(userName, password string) plugin.Plugin {
	return &LRPCAuthorization{
		UserName: userName,
		Password: password,
	}
}

func (a *LRPCAuthorization) Receive4S(pub *plugin.Context, msg *message.Message) perror.LErrorDesc {
	name := msg.MetaData.Load(NameKey)
	password := msg.MetaData.Load(PasswordKey)
	if name != a.UserName && password != a.Password {
		pub.Logger.Info("authorization failed user_name=%s password=%s", name, password)
		return ErrorFailed
	}
	return nil
}

func (a *LRPCAuthorization) Send4C(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		return nil
	}
	if a.UserName == "" && a.Password == "" {
		return nil
	}
	msg.MetaData.Store(NameKey, a.UserName)
	msg.MetaData.Store(PasswordKey, a.Password)
	return nil
}

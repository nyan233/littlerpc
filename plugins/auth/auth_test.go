package auth

import (
	"context"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuthorization(t *testing.T) {
	p := NewBasicAuth("xiaomi", "123456")
	ctx := &plugin.Context{
		PluginContext: context.Background(),
		Logger:        logger.DefaultLogger,
		EHandler:      errorhandler.DefaultErrHandler,
	}
	msg := message.New()
	assert.Nil(t, p.Send4C(ctx, msg, nil))
	assert.Nil(t, p.Receive4S(ctx, msg))
	assert.Nil(t, p.AfterReceive4C(ctx, nil, nil))
}

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/internal/pool"
	"github.com/nyan233/littlerpc/pkg/common/msgparser"
	"github.com/nyan233/littlerpc/pkg/common/msgwriter"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol/message"
	"math"
	"reflect"
	"testing"
	"time"
)

type testObject struct {
	userName string
	userId   int
}

func (t *testObject) SetUserName(ctx context.Context, userName string) error {
	t.userName = userName
	return nil
}

func (t *testObject) SetUserId(ctx context.Context, userId int) error {
	t.userId = userId
	return nil
}

func (t *testObject) GetUserId(ctx context.Context) (int, error) {
	return t.userId, nil
}

func (t *testObject) GetUserName(ctx context.Context) (string, error) {
	return t.userName, nil
}

func newTestServer(nilConn transport.ConnAdapter) (*Server, error) {
	server := &Server{}
	err := server.RegisterClass(ReflectionSource, new(LittleRpcReflection), nil)
	if err != nil {
		return nil, err
	}
	server.ctxManager = new(contextManager)
	server.connsSourceDesc.Store(nilConn, &connSourceDesc{
		Parser: msgparser.NewLRPCTrait(msgparser.NewDefaultSimpleAllocTor(), 4096),
		Writer: msgwriter.NewLRPCTrait(),
	})
	sc := &Config{}
	WithDefaultServer()(sc)
	server.eHandle = sc.ErrHandler
	server.taskPool = pool.NewTaskPool[string](sc.PoolBufferSize, sc.PoolMinSize, sc.PoolMaxSize, nil)
	server.logger = &testLogger{logger: sc.Logger}
	server.pManager = &pluginManager{plugins: sc.Plugins}
	server.config = sc
	return server, nil
}

func TestOnMessage(t *testing.T) {
	nc := &transport.NilConn{}
	server, err := newTestServer(nc)
	if err != nil {
		t.Fatal(err)
	}
	obj := reflect.ValueOf(new(testObject))
	err = server.RegisterClass("littlerpc/test/testObject", new(testObject), nil)
	if err != nil {
		t.Fatal(err)
	}
	// open debug
	server.config.Debug = true
	msg := message.New()
	for i := 0; i < obj.NumMethod(); i++ {
		msg.SetMsgType(message.Call)
		method := obj.Method(i)
		msg.SetServiceName(fmt.Sprintf("littlerpc/test/testObject.%s", obj.Type().Method(i).Name))
		for j := 1; j < method.Type().NumIn(); j++ {
			payloads, err := baseTypeGenToJson(method.Type().In(j))
			if err != nil {
				t.Fatal(err)
			}
			msg.AppendPayloads(payloads)
		}
		var bytes container.Slice[byte]
		err = message.Marshal(msg, &bytes)
		if err != nil {
			t.Fatal(err)
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					t.Fatal(err)
				}
			}()
			server.onMessage(nc, bytes)
			time.Sleep(time.Millisecond * 100)
			msg.Reset()
		}()
	}
}

func baseTypeGenToJson(typ reflect.Type) ([]byte, error) {
	switch typ.Kind() {
	case reflect.String:
		return json.Marshal(random.GenStringOnAscii(300))
	case reflect.Int64, reflect.Int, reflect.Int32:
		return json.Marshal(random.FastRandN(math.MaxUint32 / 2))
	default:
		return nil, errors.New("no match for base type")
	}
}

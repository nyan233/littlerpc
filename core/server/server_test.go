package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/metadata"
	msgparser2 "github.com/nyan233/littlerpc/core/common/msgparser"
	"github.com/nyan233/littlerpc/core/common/msgwriter"
	transport2 "github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/container"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/random"
	"github.com/nyan233/littlerpc/internal/pool"
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

func newTestServer(nilConn transport2.ConnAdapter) (*Server, error) {
	server := &Server{
		services: *container.NewRCUMap[string, *metadata.Process](),
		sources:  *container.NewRCUMap[string, *metadata.Source](),
	}
	err := server.RegisterClass(ReflectionSource, new(LittleRpcReflection), nil)
	if err != nil {
		return nil, err
	}
	server.connsSourceDesc.Store(nilConn, &connSourceDesc{
		Parser: msgparser2.NewLRPCTrait(msgparser2.NewDefaultSimpleAllocTor(), 4096),
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
	nc := &transport2.NilConn{}
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
	msg := message2.New()
	for i := 0; i < obj.NumMethod(); i++ {
		msg.SetMsgType(message2.Call)
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
		err = message2.Marshal(msg, &bytes)
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

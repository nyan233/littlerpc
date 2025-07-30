package main

import (
	"context"
	"fmt"
	"github.com/nyan233/littlerpc/core/client"
	context2 "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/common/metadata"
	"github.com/nyan233/littlerpc/core/middle/ns"
	server2 "github.com/nyan233/littlerpc/core/server"
	"github.com/nyan233/littlerpc/plugins/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zbh255/bilog"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type User struct {
	Id   int
	Name string
}

func (u *User) Reset() {
	u.Id = 0
	u.Name = ""
}

type HelloTest struct {
	count int64
	// 社区旗下的一些用户
	userMap sync.Map
	t       *testing.T
	server2.RpcServer
}

func (t *HelloTest) Setup() {
	err := t.HijackProcess("GetCount", func(stub *server2.Stub) {
		assert.NoError(t.t, stub.Write(atomic.LoadInt64(&t.count)))
		assert.NoError(t.t, stub.Write(nil))
		assert.NoError(t.t, stub.WriteErr(nil))
	})
	assert.NoError(t.t, err)
	err = t.HijackProcess("WaitSelectUserHijack", func(stub *server2.Stub) {
		var uid int
		assert.NoError(t.t, stub.Read(&uid))
		// wait
		<-stub.Done()
		user, _, err := t.SelectUser(stub.Context, uid)
		assert.NoError(t.t, stub.Write(&user))
		assert.NoError(t.t, stub.WriteErr(err))
	})
	assert.NoError(t.t, err)
}

func (t *HelloTest) GetCount(ctx *context2.Context) (int64, *User, error) {
	return atomic.LoadInt64(&t.count), nil, nil
}

func (t *HelloTest) Add(ctx *context2.Context, i int64) error {
	atomic.AddInt64(&t.count, i)
	return nil
}

func (t *HelloTest) CreateUser(ctx *context2.Context, user *User) error {
	t.userMap.Store(user.Id, *user)
	return nil
}

func (t *HelloTest) DeleteUser(ctx *context2.Context, uid int) error {
	t.userMap.Delete(uid)
	return nil
}

func (t *HelloTest) SelectUser(ctx *context2.Context, uid int) (User, bool, error) {
	u, ok := t.userMap.Load(uid)
	if ok {
		return u.(User), ok, nil
	}
	return User{}, false, nil
}

func (t *HelloTest) ModifyUser(ctx *context2.Context, uid int, user User) (bool, error) {
	_, ok := t.userMap.LoadOrStore(uid, user)
	return ok, nil
}

func (t *HelloTest) WaitSelectUser(ctx *context2.Context, uid int) (*User, error) {
	<-ctx.Done()
	user, _, err := t.SelectUser(ctx, uid)
	return &user, err
}

func (t *HelloTest) WaitSelectUserHijack(ctx *context2.Context, uid int) (*User, error) {
	return nil, nil
}

func TestServerAndClient(t *testing.T) {
	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:7878", nil))
	}()
	var (
		addrs = []string{"127.0.0.1:9999"}
	)
	// 关闭服务器烦人的日志
	logger.SetOpenLogger(false)
	baseServerOpts := []server2.Option{
		server2.WithAddressServer(addrs...),
		server2.WithStackTrace(),
		server2.WithLogger(logger.New(bilog.NewLogger(os.Stdout, bilog.PANIC,
			bilog.WithLowBuffer(0), bilog.WithTopBuffer(0)))),
		server2.WithOpenLogger(false),
		server2.WithDebug(false),
		// server2.WithMessageParserOnRead(),
		// server2.WithPlugin(pLogger.New(os.Stdout)),
	}
	baseClientOpts := []client.Option{
		client.WithNsStorage(ns.NewFixedStorage(addrs)),
		client.WithMuxConnectionNumber(16),
		client.WithStackTrace(),
	}
	testRunConfigs := []struct {
		TestName                         string
		NoAbleUsageNoTransactionProtocol bool
		ServerOptions                    []server2.Option
		ClientOptions                    []client.Option
		CallOptions                      map[string][]client.CallOption
	}{
		{
			TestName:      "TestLRPCProtocol-%s-NoMux-NonTls",
			ServerOptions: append(baseServerOpts),
			ClientOptions: append(baseClientOpts, client.WithNoMuxWriter()),
			CallOptions: map[string][]client.CallOption{
				"SelectUser": {
					client.WithCallLRPCMuxWriter(),
					client.WithCallPacker("gzip"),
				},
			},
		},
		{
			TestName:      "TestLRPCProtocol-%s-Mux-NonTls",
			ServerOptions: append(baseServerOpts),
			ClientOptions: append(baseClientOpts, client.WithMuxWriter()),
			CallOptions: map[string][]client.CallOption{
				"SelectUser": {
					client.WithCallLRPCNoMuxWriter(),
					client.WithCallPacker("gzip"),
				},
			},
		},
		{
			TestName:      "TestLRPCProtocol-%s-NoMux-Gzip-NonTls",
			ServerOptions: append(baseServerOpts),
			ClientOptions: append(baseClientOpts, client.WithNoMuxWriter(), client.WithPacker("gzip")),
		},
		{
			TestName:      "TestLRPCProtocol-%s-Mux-Gzip-NonTls",
			ServerOptions: append(baseServerOpts),
			ClientOptions: append(baseClientOpts, client.WithMuxWriter(), client.WithPacker("gzip")),
		},
		{
			TestName:                         "TestJsonRPC2-%s-SingleProtocol-NonTls",
			ServerOptions:                    append(baseServerOpts),
			ClientOptions:                    append(baseClientOpts, client.WithJsonRpc2Writer()),
			NoAbleUsageNoTransactionProtocol: true,
		},
	}
	networks := []string{"nbio_tcp", "std_tcp", "nbio_ws"}
	for _, network := range networks {
		for _, runConfig := range testRunConfigs {
			if runConfig.NoAbleUsageNoTransactionProtocol {
				switch network {
				case "nbio_tcp":
					continue
				case "std_tcp":
					continue
				}
			}
			t.Run(fmt.Sprintf(runConfig.TestName, network), func(t *testing.T) {
				testServerAndClient(t,
					append([]server2.Option{server2.WithNetwork(network)}, runConfig.ServerOptions...),
					append([]client.Option{client.WithNetWork(network)}, runConfig.ClientOptions...),
					runConfig.CallOptions)
			})
		}
	}
}

func testServerAndClient(t *testing.T, serverOpts []server2.Option, clientOpts []client.Option,
	ccSet map[string][]client.CallOption) {
	sm := metrics.NewServer()
	server := server2.New(append(serverOpts, server2.WithPlugin(sm))...)
	h := &HelloTest{
		t: t,
	}
	err := server.RegisterClass("", h, map[string]metadata.ProcessOption{
		"SelectUser": {
			SyncCall:        true,
			CompleteReUsage: true,
		},
		"CreateUser": {
			SyncCall:        true,
			CompleteReUsage: true,
		},
		"Add": {
			SyncCall:        true,
			CompleteReUsage: true,
		},
		"WaitSelectUser": {
			CompleteReUsage: false,
			UseRawGoroutine: true,
		},
	})
	assert.NoError(t, err, "server register class failed")
	go server.Service()

	defer server.Stop()

	var wg sync.WaitGroup
	// 启动多少的客户端
	nGoroutine := 10
	// 一个客户端连续发送多少次消息
	sendN := 50
	addV := 65536
	wg.Add(nGoroutine)
	cm := metrics.NewClient()
	c, err := client.New(append(clientOpts, client.WithPlugin(cm))...)
	assert.NoError(t, err, "client start failed")
	proxy := NewHelloTest(c)
	for i := 0; i < nGoroutine; i++ {
		j := i
		go func() {
			ctx := context2.Background()
			for k := 0; k < sendN; k++ {
				assert.NoError(t, proxy.Add(ctx, int64(addV)), "add failed")
				var opts []client.CallOption
				opts, _ = ccSet["CreateUser"]
				assert.NoError(t, proxy.CreateUser(ctx, &User{
					Id:   j + 100,
					Name: "Jeni",
				}, opts...), "create user failed")
				opts, _ = ccSet["SelectUser"]
				_ = h
				user, _, err := proxy.SelectUser(ctx, j+100, opts...)
				assert.NoError(t, err, "select user failed")
				assert.Equal(t, user.Name, "Jeni", "the no value")
				opts, _ = ccSet["ModifyUser"]
				_, err = proxy.ModifyUser(ctx, j+100, User{
					Id:   j + 100,
					Name: "Tony",
				}, opts...)
				assert.NoError(t, err, "modify user failed")
				opts, _ = ccSet["GetCount"]
				_, _, err = proxy.GetCount(ctx, opts...)
				assert.NoError(t, err, "get count failed")
				opts, _ = ccSet["DeleteUser"]
				assert.NoError(t, proxy.DeleteUser(ctx, j+100, opts...), "delete user failed")
				// 构造一次错误的请求
				err = c.Request2("HelloTest.DeleteUser", nil, 2, ctx, "string")
				assert.Error(t, err, "call error is equal nil")
				// 构造一次取消的请求
				ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*10)
				opts, _ = ccSet["WaitSelectUser"]
				_, err = proxy.WaitSelectUser(context2.NewContext(ctx), j+100, opts...)
				assert.NoError(t, err, "cancel request failed")
				// test hijack context
				ctx, _ = context.WithTimeout(context.Background(), time.Millisecond*10)
				_, err = proxy.WaitSelectUserHijack(context2.NewContext(ctx), 10)
				assert.NoError(t, err, "cancel request failed")
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, atomic.LoadInt64(&h.count), int64(addV*nGoroutine)*int64(sendN), "h.count no correct")

	assert.Equal(t, cm.Call.LoadFailed(), sm.Call.LoadFailed())
	//assert.Equal(t, cm.Call.LoadComplete(), sm.Call.LoadComplete())
	//assert.Equal(t, cm.Call.LoadAll(), sm.Call.LoadAll())
	//assert.Equal(t, cm.Call.LoadCount(), sm.Call.LoadCount())

	assert.Equal(t, cm.Call.LoadFailed(), int64(nGoroutine)*int64(sendN), "errCount size not correct")
}

func TestBalance(t *testing.T) {
	// 关闭服务器烦人的日志
	var (
		addrList = []string{"127.0.0.1:9090", "127.0.0.1:8080"}
		ht       = new(HelloTest)
	)
	logger.SetOpenLogger(false)
	server := server2.New(server2.WithAddressServer(addrList...),
		server2.WithOpenLogger(false))
	err := server.RegisterClass("", ht, nil)
	if err != nil {
		t.Fatal(err)
	}
	go server.Service()
	time.Sleep(time.Second)
	defer server.Stop()
	c1, err := client.New(
		client.WithNsStorage(ns.NewFixedStorage(addrList)),
		client.WithNsSchemeHash(),
	)
	if err != nil {
		t.Fatal(err)
	}
	p1 := NewHelloTest(c1)
	c2, err := client.New(
		client.WithNsStorage(ns.NewFixedStorage(addrList)),
		client.WithNsSchemeHash(),
	)
	if err != nil {
		t.Fatal(err)
	}
	p2 := NewHelloTest(c2)
	require.NoError(t, p1.Add(context2.Background(), 1024))
	require.NoError(t, p2.Add(context2.Background(), 1023))
	require.Equal(t, ht.count, int64(1024+1023))
}

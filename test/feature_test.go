package main

import (
	"context"
	"fmt"
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/common/metadata"
	server2 "github.com/nyan233/littlerpc/core/server"
	"github.com/nyan233/littlerpc/plugins/metrics"
	"github.com/stretchr/testify/assert"
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
}

func (t *HelloTest) GetCount() (int64, *User, error) {
	return atomic.LoadInt64(&t.count), nil, nil
}

func (t *HelloTest) Add(i int64) error {
	atomic.AddInt64(&t.count, i)
	return nil
}

func (t *HelloTest) CreateUser(ctx context.Context, user *User) error {
	t.userMap.Store(user.Id, *user)
	return nil
}

func (t *HelloTest) DeleteUser(ctx context.Context, uid int) error {
	t.userMap.Delete(uid)
	return nil
}

func (t *HelloTest) SelectUser(ctx context.Context, uid int) (User, bool, error) {
	u, ok := t.userMap.Load(uid)
	if ok {
		return u.(User), ok, nil
	}
	return User{}, false, nil
}

func (t *HelloTest) ModifyUser(ctx context.Context, uid int, user User) (bool, error) {
	_, ok := t.userMap.LoadOrStore(uid, user)
	return ok, nil
}

func (t *HelloTest) WaitSelectUser(ctx context.Context, uid int) (*User, error) {
	<-ctx.Done()
	user, _, err := t.SelectUser(ctx, uid)
	return &user, err
}

func TestServerAndClient(t *testing.T) {
	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:7878", nil))
	}()
	// 关闭服务器烦人的日志
	logger.SetOpenLogger(false)
	baseServerOpts := []server2.Option{
		server2.WithAddressServer(":1234"),
		server2.WithStackTrace(),
		server2.WithLogger(logger.New(bilog.NewLogger(os.Stdout, bilog.PANIC))),
		server2.WithOpenLogger(false),
		server2.WithDebug(false),
		//server2.WithPlugin(pLogger.New(os.Stdout)),
	}
	baseClientOpts := []client.Option{
		client.WithAddress(":1234"),
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
	h := &HelloTest{}
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
	nGoroutine := 50
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
			for k := 0; k < sendN; k++ {
				assert.NoError(t, proxy.Add(int64(addV)), "add failed")
				var opts []client.CallOption
				opts, _ = ccSet["CreateUser"]
				assert.NoError(t, proxy.CreateUser(context.Background(), &User{
					Id:   j + 100,
					Name: "Jeni",
				}, opts...), "create user failed")
				opts, _ = ccSet["SelectUser"]
				user, _, err := proxy.SelectUser(context.Background(), j+100, opts...)
				assert.NoError(t, err, "select user failed")
				assert.Equal(t, user.Name, "Jeni", "the no value")
				opts, _ = ccSet["ModifyUser"]
				_, err = proxy.ModifyUser(context.Background(), j+100, User{
					Id:   j + 100,
					Name: "Tony",
				}, opts...)
				assert.NoError(t, err, "modify user failed")
				opts, _ = ccSet["GetCount"]
				_, _, err = proxy.GetCount(opts...)
				assert.NoError(t, err, "get count failed")
				opts, _ = ccSet["DeleteUser"]
				assert.NoError(t, proxy.DeleteUser(context.Background(), j+100, opts...), "delete user failed")
				// 构造一次错误的请求
				_, err = c.Call("HelloTest.DeleteUser", nil, context.Background(), "string")
				assert.Error(t, err, "call error is equal nil")
				// 构造一次取消的请求
				ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*10)
				var rep User
				opts, _ = ccSet["WaitSelectUser"]
				assert.NoError(t, c.Request("HelloTest.WaitSelectUser", ctx, j+100, &rep, opts...), "cancel request failed")
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
	logger.SetOpenLogger(false)
	server := server2.New(server2.WithAddressServer("127.0.0.1:9090", "127.0.0.1:8080"),
		server2.WithOpenLogger(false))
	err := server.RegisterClass("", new(HelloTest), nil)
	if err != nil {
		t.Fatal(err)
	}
	go server.Service()
	defer server.Stop()
	c1, err := client.New(
		client.WithBalance("roundRobin"),
		client.WithResolver("live", "live://127.0.0.1:8080;127.0.0.1:9090"),
	)
	if err != nil {
		t.Fatal(err)
	}
	p1 := NewHelloTest(c1)
	c2, err := client.New(
		client.WithBalance("roundRobin"),
		client.WithResolver("live", "live://127.0.0.1:8080;127.0.0.1:9090"),
	)
	if err != nil {
		t.Fatal(err)
	}
	p2 := NewHelloTest(c2)
	p1.Add(1024)
	p2.Add(1023)
}

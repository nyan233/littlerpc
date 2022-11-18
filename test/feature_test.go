package main

import (
	"context"
	lclient "github.com/nyan233/littlerpc/client"
	"github.com/nyan233/littlerpc/pkg/common/logger"
	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/plugins/metrics"
	lserver "github.com/nyan233/littlerpc/server"
	"log"
	"net/http"
	_ "net/http/pprof"
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
	serverOpts := []lserver.Option{
		lserver.WithAddressServer(":1234"),
		lserver.WithOpenLogger(false),
		//lserver.WithDebug(true),
	}
	clientOpts := []lclient.Option{
		lclient.WithAddress(":1234"),
		lclient.WithMuxConnectionNumber(16),
		lclient.WithPlugin(&metrics.ClientMetricsPlugin{}),
	}
	t.Run("TestLRPCNoMuxProtocolNonTls", func(t *testing.T) {
		cOpts := clientOpts
		cOpts = append(cOpts, lclient.WithMuxWriter())
		testServerAndClient(t, serverOpts, cOpts)
	})
	//t.Run("TestLRPCProtocolGzipNonTls", func(t *testing.T) {
	//	cOpts := clientOpts
	//	cOpts = append(cOpts, lclient.WithUseMux(false))
	//	cOpts = append(cOpts, lclient.WithPacker("gzip"))
	//	testServerAndClient(t, serverOpts, clientOpts)
	//})
	//t.Run("TestLRPCMuxProtocolNonTls", func(t *testing.T) {
	//	cOpts := clientOpts
	//	cOpts = append(cOpts, lclient.WithUseMux(true))
	//	testServerAndClient(t, serverOpts, clientOpts)
	//})
	//t.Run("TestJsonRPC2SingleProtocolNonTls", func(t *testing.T) {
	//	cOpts := clientOpts
	//	cOpts = append(cOpts, lclient.WithUseMux(false))
	//	cOpts = append(cOpts, lclient.WithWriter(msgwriter.Manager.Get(jsonrpc2.Header)))
	//	cOpts = append(cOpts, lclient.WithProtocol("nbio_ws"))
	//	sOpts := serverOpts
	//	sOpts = append(sOpts, lserver.WithNetwork("nbio_ws"))
	//	testServerAndClient(t, sOpts, cOpts)
	//})
}

func testServerAndClient(t *testing.T, serverOpts []lserver.Option, clientOpts []lclient.Option) {
	server := lserver.New(serverOpts...)
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
	if err != nil {
		t.Fatal(err)
	}
	err = server.Start()
	if err != nil {
		t.Fatal(err)
	}

	defer server.Stop()

	var wg sync.WaitGroup
	// 启动多少的客户端
	nGoroutine := 1000
	// 一个客户端连续发送多少次消息
	sendN := 50
	addV := 65536
	wg.Add(nGoroutine)
	client, err := lclient.New(clientOpts...)
	if err != nil {
		t.Fatal(err)
	}
	proxy := NewHelloTestProxy(client)
	for i := 0; i < nGoroutine; i++ {
		j := i
		go func() {
			for i := 0; i < sendN; i++ {
				_ = proxy.Add(int64(addV))
				_ = proxy.CreateUser(context.Background(), &User{
					Id:   j + 100,
					Name: "Jeni",
				})
				user, _, err := proxy.SelectUser(context.Background(), j+100)
				if err != nil {
					panic(err)
				}
				if user.Name != "Jeni" {
					panic(interface{}("the no value"))
				}
				_, _ = proxy.ModifyUser(context.Background(), j+100, User{
					Id:   j + 100,
					Name: "Tony",
				})
				_, _, err = proxy.GetCount()
				if err != nil {
					t.Error(err)
				}
				_ = proxy.DeleteUser(context.Background(), j+100)
				pp := proxy.(*HelloTestProxy)
				// 构造一次错误的请求
				_, err = pp.Call("HelloTest.DeleteUser", context.Background(), "string")
				if err == nil {
					t.Error("call error is equal nil")
				}
				// 构造一次取消的请求
				ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*10)
				var rep User
				err = pp.Request("HelloTest.WaitSelectUser", ctx, j+100, &rep)
				if err != nil {
					t.Log(err)
				}
			}
			wg.Done()
		}()
	}
	go func() {
		for {
			time.Sleep(time.Second * 10)
			t.Log(metrics.ServerCallMetrics.LoadCount())
			t.Log(metrics.ClientCallMetrics.LoadCount())
			t.Log(proxy.GetCount())
			t.Log(atomic.LoadInt64(&h.count))
		}
	}()
	wg.Wait()
	if atomic.LoadInt64(&h.count) != int64(addV*nGoroutine)*int64(sendN) {
		t.Fatal("h.count no correct")
	}
	if metrics.ClientCallMetrics.LoadFailed() != int64(nGoroutine)*int64(sendN) {
		t.Fatal("errCount size not correct")
	}
}

func TestBalance(t *testing.T) {
	// 关闭服务器烦人的日志
	logger.SetOpenLogger(false)
	server := lserver.New(lserver.WithAddressServer("127.0.0.1:9090", "127.0.0.1:8080"),
		lserver.WithOpenLogger(false))
	err := server.RegisterClass("", new(HelloTest), nil)
	if err != nil {
		t.Fatal(err)
	}
	err = server.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	c1, err := lclient.New(
		lclient.WithBalance("roundRobin"),
		lclient.WithResolver("live", "live://127.0.0.1:8080;127.0.0.1:9090"),
	)
	if err != nil {
		t.Fatal(err)
	}
	p1 := NewHelloTestProxy(c1)
	c2, err := lclient.New(
		lclient.WithBalance("roundRobin"),
		lclient.WithResolver("live", "live://127.0.0.1:8080;127.0.0.1:9090"),
	)
	if err != nil {
		t.Fatal(err)
	}
	p2 := NewHelloTestProxy(c2)
	p1.Add(1024)
	p2.Add(1023)
}

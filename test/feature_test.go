package main

import (
	lclient "github.com/nyan233/littlerpc/client"
	"github.com/nyan233/littlerpc/pkg/common"
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

func (t *HelloTest) CreateUser(user *User) error {
	t.userMap.Store(user.Id, *user)
	return nil
}

func (t *HelloTest) DeleteUser(uid int) error {
	t.userMap.Delete(uid)
	return nil
}

func (t *HelloTest) SelectUser(uid int) (User, bool, error) {
	u, ok := t.userMap.Load(uid)
	if ok {
		return u.(User), ok, nil
	}
	return User{}, false, nil
}

func (t *HelloTest) ModifyUser(uid int, user User) (bool, error) {
	_, ok := t.userMap.LoadOrStore(uid, user)
	return ok, nil
}

func TestNoTlsServerAndClient(t *testing.T) {
	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:7878", nil))
	}()
	// 关闭服务器烦人的日志
	common.SetOpenLogger(false)
	server := lserver.New(
		lserver.WithAddressServer(":1234"),
		lserver.WithTransProtocol("nbio_tcp"),
		lserver.WithServerEncoder("gzip"),
		lserver.WithOpenLogger(false),
	)
	h := &HelloTest{}
	err := server.RegisterClass(h, map[string]common.MethodOption{
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
	// 统计触发错误的次数
	var errCount int64
	addV := 65536
	wg.Add(nGoroutine)
	client, err := lclient.New(
		lclient.WithCallOnErr(func(err error) { atomic.AddInt64(&errCount, 1) }),
		lclient.WithAddressClient(":1234"),
		lclient.WithClientCodec("json"),
		//lclient.WithClientEncoder("gzip"),
		lclient.WithProtocol("nbio_tcp"),
		lclient.WithMuxConnectionNumber(64),
		lclient.WithUseMux(true),
		lclient.WithPlugin(&metrics.ClientMetricsPlugin{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	proxy := NewHelloTestProxy(client)
	for i := 0; i < nGoroutine; i++ {
		j := i
		go func() {
			for i := 0; i < sendN; i++ {
				_ = proxy.Add(int64(addV))
				_ = proxy.CreateUser(&User{
					Id:   j + 100,
					Name: "Jeni",
				})
				user, _, err := proxy.SelectUser(j + 100)
				if err != nil {
					panic(err)
				}
				if user.Name != "Jeni" {
					panic(interface{}("the no value"))
				}
				_, _ = proxy.ModifyUser(j+100, User{
					Id:   j + 100,
					Name: "Tony",
				})
				_, _, err = proxy.GetCount()
				if err != nil {
					t.Error(err)
				}
				_ = proxy.DeleteUser(j + 100)
				pp := proxy.(*HelloTestProxy)
				// 构造一次错误的请求
				_, err = pp.Call("HelloTest.DeleteUser", "string")
				if err == nil {
					t.Error("call error is equal nil")
				}
				if err != nil {
					atomic.AddInt64(&errCount, 1)
				}
			}
			wg.Done()
		}()
	}
	go func() {
		for {
			time.Sleep(time.Second * 1000)
			t.Log(metrics.ServerCallMetrics.LoadCount())
			t.Log(metrics.ClientCallMetrics.LoadCount())
			t.Log(proxy.GetCount())
			t.Log(atomic.LoadInt64(&errCount))
			t.Log(atomic.LoadInt64(&h.count))
		}
	}()
	wg.Wait()
	if atomic.LoadInt64(&h.count) != int64(addV*nGoroutine)*int64(sendN) {
		t.Fatal("h.count no correct")
	}
	if atomic.LoadInt64(&errCount) != int64(nGoroutine)*int64(sendN) {
		t.Fatal("errCount size not correct")
	}
}

func TestBalance(t *testing.T) {
	// 关闭服务器烦人的日志
	common.SetOpenLogger(false)
	server := lserver.New(lserver.WithAddressServer("127.0.0.1:9090", "127.0.0.1:8080"),
		lserver.WithOpenLogger(false))
	err := server.RegisterClass(new(HelloTest), nil)
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

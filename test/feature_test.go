package test

import (
	lclient "github.com/nyan233/littlerpc/impl/client"
	"github.com/nyan233/littlerpc/impl/common"
	lserver "github.com/nyan233/littlerpc/impl/server"
	"math"
	"sync"
	"sync/atomic"
	"testing"
)

type User struct {
	Id   int
	Name string
}

type HelloTest struct {
	count int64
	// 社区旗下的一些用户
	userMap sync.Map
}

func (t *HelloTest) GetCount() (int64,*User,error) {
	return atomic.LoadInt64(&t.count),nil,nil
}

func (t *HelloTest) Add(i int64) error {
	atomic.AddInt64(&t.count, i)
	return nil
}

func (t *HelloTest) CreateUser(user User) error {
	t.userMap.Store(user.Id, user)
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


func TestNoTlsConnect(t *testing.T) {
	// 关闭服务器烦人的日志
	common.SetOpenLogger(false)
	server := lserver.NewServer(
		lserver.WithAddressServer(":1234"),
		lserver.WithServerEncoder("gzip"),
		lserver.WithOpenLogger(false),
	)
	h := &HelloTest{}
	err := server.Elem(h)
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
	nGoroutine := 100
	// 一个客户端连续发送多少次消息
	sendN := 5
	// 统计触发错误的次数
	var errCount int64
	addV := 65536
	wg.Add(nGoroutine)
	for i := 0; i < nGoroutine; i++ {
		j := i
		go func() {
			client, err := lclient.NewClient(
				lclient.WithCallOnErr(func(err error) { atomic.AddInt64(&errCount, 1) }),
				lclient.WithAddressClient(":1234"),
				lclient.WithClientCodec("json"),
				lclient.WithClientEncoder("gzip"),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			proxy := NewHelloTestProxy(client)
			for i := 0; i < sendN; i++ {
				_ = proxy.Add(int64(addV))
				_ = proxy.CreateUser(User{
					Id:   j + 100,
					Name: "Jeni",
				})
				user, ok, _ := proxy.SelectUser(j + 100)
				if !ok {
					panic("the no value")
				}
				if user.Name != "Jeni" {
					panic("the no value")
				}
				_, _ = proxy.ModifyUser(j+100, User{
					Id:   j + 100,
					Name: "Tony",
				})
				_, _, err := proxy.GetCount()
				if err != nil {
					t.Error(err)
				}
				_ = proxy.DeleteUser(j + 100)
				pp := proxy.(*HelloTestProxy)
				// 构造一次错误的请求
				_, err = pp.Call("HelloTest.DeleteUser", "string")
				if err != nil {
					atomic.AddInt64(&errCount,1)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if atomic.LoadInt64(&h.count) != int64(addV*nGoroutine) * int64(sendN) {
		t.Fatal("h.count no correct")
	}
	if atomic.LoadInt64(&errCount) != int64(nGoroutine) * int64(sendN) {
		t.Fatal("errCount size not correct")
	}
}

func TestTlsConnect(t *testing.T) {

}

func TestBalance(t *testing.T) {
	server := lserver.NewServer(lserver.WithAddressServer("127.0.0.1:9090", "127.0.0.1:8080"))
	err := server.Elem(new(HelloTest))
	if err != nil {
		t.Fatal(err)
	}
	err = server.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	err = lclient.OpenBalance("live", "live://127.0.0.1:8080;127.0.0.1:9090", math.MaxInt64)
	if err != nil {
		t.Fatal(err)
	}
	c1, err := lclient.NewClient(lclient.WithBalance("roundRobin"))
	if err != nil {
		t.Fatal(err)
	}
	p1 := NewHelloTestProxy(c1)
	c2, err := lclient.NewClient(lclient.WithBalance("roundRobin"))
	if err != nil {
		t.Fatal(err)
	}
	p2 := NewHelloTestProxy(c2)
	p1.Add(1024)
	p2.Add(1023)
}

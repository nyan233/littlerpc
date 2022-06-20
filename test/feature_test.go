package test

import (
	lclient "github.com/nyan233/littlerpc/impl/client"
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

func (t *HelloTest) Add(i int64) error {
	atomic.AddInt64(&t.count, i)
	return nil
}

func (t *HelloTest) CreateUser(user User) error{
	t.userMap.Store(user.Id, user)
	return nil
}

func (t *HelloTest) DeleteUser(uid int) error{
	t.userMap.Delete(uid)
	return nil
}

func (t *HelloTest) SelectUser(uid int) (User, bool, error) {
	u, ok := t.userMap.Load(uid)
	if ok {
		return u.(User), ok,nil
	}
	return User{}, false,nil
}

func (t *HelloTest) ModifyUser(uid int, user User) (bool,error) {
	_, ok := t.userMap.LoadOrStore(uid, user)
	return ok,nil
}

type HelloTestProxy struct {
	*lclient.Client
}

func NewHelloTestProxy(client *lclient.Client) *HelloTestProxy {
	proxy := &HelloTestProxy{}
	err := client.BindFunc(proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (proxy HelloTestProxy) Add(i int64) error {
	_, err := proxy.Call("HelloTest.Add", i)
	return err
}

func (proxy HelloTestProxy) CreateUser(user User) error{
	_, err := proxy.Call("HelloTest.CreateUser", user)
	return err
}

func (proxy HelloTestProxy) DeleteUser(uid int) error {
	_, err := proxy.Call("HelloTest.DeleteUser", uid)
	return err
}

func (proxy HelloTestProxy) SelectUser(uid int) (User, bool, error) {
	inter, err := proxy.Call("HelloTest.SelectUser", uid)
	r0 := inter[0].(User)
	r1 := inter[1].(bool)
	return r0, r1,err
}

func (proxy HelloTestProxy) ModifyUser(uid int, user User) (bool,error) {
	inter, err := proxy.Call("HelloTest.ModifyUser", uid, user)
	r0 := inter[0].(bool)
	return r0,err
}

func TestNoTlsConnect(t *testing.T) {
	server := lserver.NewServer(
		lserver.WithAddressServer(":1234"),
		lserver.WithServerEncoder("gzip"),
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
	nGoroutine := 100
	// 统计触发错误的次数
	var errCount int64
	addV := 65536
	wg.Add(nGoroutine)
	for i := 0; i < nGoroutine; i++ {
		j := i
		go func() {
			client,err := lclient.NewClient(
				lclient.WithCallOnErr(func(err error) {atomic.AddInt64(&errCount, 1)}),
				lclient.WithAddressClient(":1234"),
				lclient.WithClientEncoder("gzip"),
				)
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			proxy := NewHelloTestProxy(client)
			for i := 0; i < 5; i++ {
				_ = proxy.Add(int64(addV))
				_ = proxy.CreateUser(User{
					Id:   j + 100,
					Name: "Jeni",
				})
				user, ok,_ := proxy.SelectUser(j + 100)
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
				_ = proxy.DeleteUser(j + 100)
				// 构造一次错误的请求
				_, _ = proxy.Call("HelloTest.DeleteUser", "string")
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if atomic.LoadInt64(&h.count) != int64(addV*nGoroutine) {
		t.Fatal("h.count no correct")
	}
	if atomic.LoadInt64(&errCount) != int64(nGoroutine) {
		t.Fatal("errCount size not correct")
	}
}

func TestTlsConnect(t *testing.T) {

}

func TestBalance(t *testing.T) {
	server := lserver.NewServer(lserver.WithAddressServer("127.0.0.1:9090","127.0.0.1:8080"))
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
	c1,err := lclient.NewClient(lclient.WithBalance("roundRobin"))
	if err != nil {
		t.Fatal(err)
	}
	p1 := NewHelloTestProxy(c1)
	c2,err := lclient.NewClient(lclient.WithBalance("roundRobin"))
	if err != nil {
		t.Fatal(err)
	}
	p2 := NewHelloTestProxy(c2)
	p1.Add(1024)
	p2.Add(1023)
}
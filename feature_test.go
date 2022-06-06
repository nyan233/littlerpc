package littlerpc

import (
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

func (t *HelloTest) Add(i int64) {
	atomic.AddInt64(&t.count,i)
}

func (t *HelloTest) CreateUser(user User) {
	t.userMap.Store(user.Id,user)
}

func (t *HelloTest) DeleteUser(uid int) {
	t.userMap.Delete(uid)
}

func (t *HelloTest) SelectUser(uid int) (User,bool) {
	u,ok := t.userMap.Load(uid)
	if ok {
		return u.(User),ok
	}
	return User{},false
}

func (t *HelloTest) ModifyUser(uid int,user User) bool {
	_,ok := t.userMap.LoadOrStore(uid,user)
	return ok
}

type HelloTestProxy struct {
	*Client
}

func NewHelloTestProxy(client *Client) *HelloTestProxy {
	proxy := &HelloTestProxy{}
	err := client.BindFunc(proxy)
	if err != nil {
		panic(err)
	}
	proxy.Client = client
	return proxy
}

func (proxy HelloTestProxy) Add(i int64) {
	_, _ = proxy.Call("Add", i)
	return
}

func (proxy HelloTestProxy) CreateUser(user User) {
	_, _ = proxy.Call("CreateUser", user)
	return
}

func (proxy HelloTestProxy) DeleteUser(uid int) {
	_, _ = proxy.Call("DeleteUser", uid)
	return
}

func (proxy HelloTestProxy) SelectUser(uid int) (User, bool) {
	inter, _ := proxy.Call("SelectUser", uid)
	r0 := inter[0].(User)
	r1 := inter[1].(bool)
	return r0, r1
}

func (proxy HelloTestProxy) ModifyUser(uid int, user User) bool {
	inter, _ := proxy.Call("ModifyUser", uid, user)
	r0 := inter[0].(bool)
	return r0
}



func TestNoTlsConnect(t *testing.T) {
	server := NewServer(WithAddressServer(":1234"))
	h := &HelloTest{}
	err := server.Elem(h)
	if err != nil {
		t.Fatal(err)
	}
	err = server.Start()
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	nGoroutine := 1
	addV := 65536
	wg.Add(nGoroutine)
	for i := 0; i < nGoroutine; i++ {
		j := i
		go func() {
			client := NewClient(WithCallOnErr(func(err error) {
				panic(err)
			}),WithAddressClient(":1234"))
			proxy := NewHelloTestProxy(client)
			proxy.Add(int64(addV))
			proxy.CreateUser(User{
				Id:   j + 100,
				Name: "Jeni",
			})
			user,ok := proxy.SelectUser(j + 100)
			if !ok {
				panic("the no value")
			}
			if user.Name != "Jeni" {
				panic("the no value")
			}
			proxy.ModifyUser(j + 100,User{
				Id:   j + 100,
				Name: "Tony",
			})
			proxy.DeleteUser(j + 100)
			// 构造一次错误的请求
			_, err = proxy.Call("DeleteUser", "string")
			wg.Done()
		}()
	}
	wg.Wait()
	if h.count != int64(addV * nGoroutine) {
		t.Fatal("h.count no correct")
	}
}

package main

import (
	_ "net/http/pprof"
)

//func main() {
//	go func() {
//		log.Println(http.ListenAndServe(":6060", nil))
//	}()
//	server := server.NewServer(server.WithAddressServer(":1234"))
//	i1 := new(HelloServer1)
//	i2 := new(HelloServer2)
//	err := server.Elem(i1)
//	if err != nil {
//		panic(err)
//	}
//	err = server.Elem(i2)
//	if err != nil {
//		panic(err)
//	}
//	_ = server.Start()
//	defer server.Stop()
//	time.Sleep(time.Second * 100000)
//	client1, err := client.NewClient(client.WithAddressClient(":1234"))
//	if err != nil {
//		panic(err)
//	}
//	ci1 := NewHelloServer1Proxy(client1)
//	client2, err := client.NewClient(client.WithAddressClient(":1234"))
//	if err != nil {
//		panic(err)
//	}
//	ci2 := NewHelloServer2Proxy(client2)
//	println(ci1.Hello())
//	_ = ci2.Init("my name is server 2")
//	println(ci2.Hello())
//}

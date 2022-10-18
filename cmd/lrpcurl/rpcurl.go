package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nyan233/littlerpc/client"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/utils/convert"
	"github.com/nyan233/littlerpc/server"
	"strings"
)

type MethodTable = server.MethodTable
type ArgumentType = server.ArgumentType

var (
	source = flag.String("source", "", "资源的描述,Example: 127.0.0.1:9090")
	option = flag.String("option", "get_all_instance", "操作(get_all_instance | get_arg_type)")
	target = flag.String("target", "Hello.Hello", "调用的目标: InstanceName.MethodName")
	call   = flag.String("call", "null", "调用传递的参数: [100,\"hh\"]")
)

func main() {
	flag.Parse()
	common.SetOpenLogger(false)
	c := dial()
	proxy := NewLittleRpcReflectionProxy(c)
	switch *option {
	case "get_all_instance":
		getAllInstance(proxy)
	case "get_arg_type":
		getArgType(proxy)
	case "get_method_table":
		getMethodTable(proxy)
	case "call_func":
		callFunc(c)
	}
}

func dial() *client.Client {
	c, err := client.New(
		client.WithCustomLoggerClient(common.NilLogger),
		client.WithUseMux(false),
		client.WithMuxConnection(false),
		client.WithProtocol("std_tcp"),
		client.WithAddressClient(*source),
	)
	*call = strings.TrimPrefix(*call, "\xef\xbb\xbf")
	if err != nil {
		panic(err)
	}
	return c
}

func getAllInstance(proxy LittleRpcReflectionInterface) {
	instance, err := proxy.AllInstance()
	if err != nil {
		panic(err)
	}
	iBytes, err := json.Marshal(instance)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(iBytes))
}

func getArgType(proxy LittleRpcReflectionInterface) {
	var args []string
	err := json.Unmarshal(convert.StringToBytes(*call), &args)
	if err != nil {
		panic(err)
	}
	argumentType, err := proxy.MethodArgumentType(args[0], args[1])
	if err != nil {
		panic(err)
	}
	aBytes, err := json.Marshal(argumentType)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(aBytes))
}

func getMethodTable(proxy LittleRpcReflectionInterface) {
	var args []string
	err := json.Unmarshal(convert.StringToBytes(*call), &args)
	if err != nil {
		panic(err)
	}
	table, err := proxy.MethodTable(args[0])
	if err != nil {
		panic(err)
	}
	tBytes, err := json.Marshal(table)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(tBytes))
}

func callFunc(c *client.Client) {
	var args []interface{}
	err := json.Unmarshal(convert.StringToBytes(*call), &args)
	if err != nil {
		panic(err)
	}
	reply, err := c.RawCall(*target, args...)
	if err != nil {
		panic(err)
	}
	replyBytes, err := json.Marshal(reply)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(replyBytes))
}

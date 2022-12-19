package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nyan233/littlerpc/cmd/lrpcurl/proxy"
	client2 "github.com/nyan233/littlerpc/core/client"
	logger2 "github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/server"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"io"
	"log"
	"os"
	"strings"
	"unsafe"
)

type Caller interface {
	RawCall(service string, args ...interface{}) (reps []interface{}, err error)
}

type OutType string

const (
	FormatJson OutType = "format_json"
	Json       OutType = "json"
	Text       OutType = "text"
)

const (
	GetAllSupportOption = "get_all_support_option"
	GetAllInstance      = "get_all_instance"
	GetAllCodec         = "get_all_codec"
	GetAllPacker        = "get_all_packer"
	GetMethodTable      = "get_method_table"
	GetArgumentType     = "get_argument_type"
	CallFunc            = "call_func"
)

var (
	allSupportOption = []string{
		GetAllCodec, GetAllPacker, GetAllInstance, GetMethodTable, GetArgumentType, CallFunc,
	}
)

var (
	serverAddr = flag.String("address", "127.0.0.1:9090", "服务器地址,Example: 127.0.0.1:9090")
	source     = flag.String("source", server.ReflectionSource, "资源的名称,注册方法时指定的实例名称")
	option     = flag.String("option", "get_all_instance", "操作(get_all_instance | get_arg_type)")
	service    = flag.String("service", "Hello.Hello", "调用的目标: InstanceName.MethodName")
	outType    = flag.String("out_type", string(FormatJson), "输出的信息的格式(format_json/json/text)")
	call       = flag.String("call", "null", "调用传递的参数(不包括context/stream): [100,\"hh\"]")
)

func main() {
	flag.Parse()
	logger2.SetOpenLogger(false)
	c := dial()
	parserOption(proxy.NewLittleRpcReflection(c), c)
}

func parserOption(proxy proxy.LittleRpcReflectionProxy, caller Caller) {
	switch *option {
	case GetAllSupportOption:
		getAllSupportOption(OutType(*outType), os.Stdout)
	case GetAllCodec:
		getAllCodec(proxy, OutType(*option), os.Stdout)
	case GetAllPacker:
		getAllPacker(proxy, OutType(*option), os.Stdout)
	case GetAllInstance:
		getAllInstance(proxy, OutType(*outType), os.Stdout)
	case GetArgumentType:
		getArgType(proxy, *service, OutType(*outType), os.Stdout)
	case GetMethodTable:
		getMethodTable(proxy, *source, OutType(*outType), os.Stdout)
	case CallFunc:
		var rawArgs []json.RawMessage
		err := json.Unmarshal(convert.StringToBytes(*call), &rawArgs)
		if err != nil {
			log.Fatalln(err)
		}
		callFunc(caller, *service, *(*[][]byte)(unsafe.Pointer(&rawArgs)), OutType(*outType), os.Stdout)
	}
}

func dial() *client2.Client {
	c, err := client2.New(
		client2.WithCustomLogger(logger2.NilLogger{}),
		client2.WithNoMuxWriter(),
		client2.WithMuxConnection(false),
		client2.WithProtocol("std_tcp"),
		client2.WithAddress(*serverAddr),
	)
	*call = strings.TrimPrefix(*call, "\xef\xbb\xbf")
	if err != nil {
		panic(err)
	}
	return c
}

func getAllSupportOption(ot OutType, w io.Writer) {
	switch ot {
	case Text:
		for _, option := range allSupportOption {
			_, _ = fmt.Fprintln(w, option)
		}
	default:
		anyOutFromJson(allSupportOption, ot, w)
	}
}

func getAllInstance(proxy proxy.LittleRpcReflectionProxy, ot OutType, w io.Writer) {
	instances, err := proxy.AllInstance()
	if err != nil {
		log.Fatalln(err)
	}
	switch ot {
	case Text:
		for k, v := range instances {
			_, _ = fmt.Fprintf(w, "%s --> %s\n", k, v)
		}
	default:
		anyOutFromJson(instances, ot, w)
	}
}

func getArgType(proxy proxy.LittleRpcReflectionProxy, service string, ot OutType, w io.Writer) {
	argType, err := proxy.MethodArgumentType(service)
	if err != nil {
		log.Fatalln(err)
	}
	switch ot {
	case Text:
		if argType == nil || len(argType) == 0 {
			return
		}
		for _, v := range argType {
			_, _ = fmt.Fprintln(w, v.TypeName)
		}
	default:
		anyOutFromJson(argType, ot, w)
	}
}

func getMethodTable(proxy proxy.LittleRpcReflectionProxy, sourceName string, ot OutType, w io.Writer) {
	tab, err := proxy.MethodTable(sourceName)
	if err != nil {
		log.Fatalln(err)
	}
	if tab == nil {
		return
	}
	switch ot {
	case Text:
		break
	default:
		anyOutFromJson(tab, ot, w)
	}
}

func getAllPacker(proxy proxy.LittleRpcReflectionProxy, ot OutType, w io.Writer) {
	packers, err := proxy.AllPacker()
	if err != nil {
		log.Fatalln(err)
	}
	switch ot {
	case Text:
		if packers == nil {
			return
		}
		for _, packer := range packers {
			_, _ = fmt.Fprintln(w, packer)
		}
	default:
		anyOutFromJson(packers, ot, w)
	}
}

func getAllCodec(proxy proxy.LittleRpcReflectionProxy, ot OutType, w io.Writer) {
	codecs, err := proxy.AllCodec()
	if err != nil {
		log.Fatalln(err)
	}
	switch ot {
	case Text:
		if codecs == nil {
			return
		}
		for _, codec := range codecs {
			_, _ = fmt.Fprintln(w, codec)
		}
	default:
		anyOutFromJson(codecs, ot, w)
	}
}

func anyOutFromJson(data any, ot OutType, w io.Writer) {
	switch ot {
	case FormatJson:
		bytes, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			log.Fatalln(err)
		}
		_, _ = fmt.Fprintln(w, string(bytes))
	case Json:
		bytes, err := json.Marshal(data)
		if err != nil {
			log.Fatalln(err)
		}
		_, _ = fmt.Fprintln(w, string(bytes))
	default:
		_, err := w.Write([]byte("invalid output format"))
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func callFunc(c Caller, service string, argsBytes [][]byte, ot OutType, w io.Writer) {
	args := make([]interface{}, 0, 8)
	for k, rawArg := range argsBytes {
		if len(args) == k {
			args = append(args, nil)
		}
		err := json.Unmarshal(rawArg, &args[k])
		if err != nil {
			log.Fatalln(err)
		}
	}
	reps, err := c.RawCall(service, args...)
	reps = append(reps, err)
	switch ot {
	case Text:
		for _, rep := range reps {
			bytes, err := json.Marshal(rep)
			if err != nil {
				log.Fatalln(err)
			}
			_, _ = fmt.Fprintln(w, string(bytes))
		}
	default:
		anyOutFromJson(reps, ot, w)
	}
}

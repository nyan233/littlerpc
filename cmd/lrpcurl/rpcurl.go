package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nyan233/littlerpc/cmd/lrpcurl/proxy"
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/server"
	"github.com/nyan233/littlerpc/core/utils/convert"
	flag "github.com/spf13/pflag"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unsafe"
)

type Caller interface {
	RawCall(service string, opts []client.CallOption, args ...interface{}) (reps []interface{}, err error)
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
	serverAddr = flag.StringP("address", "a", "127.0.0.1:9090", "服务器地址,Example: 127.0.0.1:9090")
	source     = flag.StringP("source", "i", server.ReflectionSource, "资源的名称,注册方法时指定的实例名称")
	option     = flag.StringP("option", "o", "get_all_instance", "操作, get_all_support_option -> 获取所有支持的操作了解更多")
	service    = flag.StringP("service", "s", "Hello.Hello", "调用的目标: InstanceName.MethodName")
	outType    = flag.StringP("out_type", "t", string(FormatJson), "输出的信息的格式(format_json/json/text)")
	call       = flag.StringP("call", "c", "null", "调用传递的参数(不包括context/stream): [100,\"hh\"]")
)

func main() {
	flag.Parse()
	logger.SetOpenLogger(false)
	c := dial()
	defer c.Close()
	ctx, cancel := context.WithCancel(context.Background())
	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		select {
		case <-channel:
			cancel()
		}
	}()
	parserOption(ctx, proxy.NewLittleRpcReflection(c), c)
}

func parserOption(ctx context.Context, proxy proxy.LittleRpcReflectionProxy, caller Caller) {
	switch *option {
	case GetAllSupportOption:
		getAllSupportOption(OutType(*outType), os.Stdout)
	case GetAllCodec:
		getAllCodec(ctx, proxy, OutType(*outType), os.Stdout)
	case GetAllPacker:
		getAllPacker(ctx, proxy, OutType(*outType), os.Stdout)
	case GetAllInstance:
		getAllInstance(ctx, proxy, OutType(*outType), os.Stdout)
	case GetArgumentType:
		getArgType(ctx, proxy, *service, OutType(*outType), os.Stdout)
	case GetMethodTable:
		getMethodTable(ctx, proxy, *source, OutType(*outType), os.Stdout)
	case CallFunc:
		var rawArgs []json.RawMessage
		err := json.Unmarshal(convert.StringToBytes(*call), &rawArgs)
		if err != nil {
			log.Fatalln(err)
		}
		callFunc(ctx, caller, *service, *(*[][]byte)(unsafe.Pointer(&rawArgs)), OutType(*outType), os.Stdout)
	default:
		log.Fatalln("no implement option")
	}
}

func dial() *client.Client {
	c, err := client.New(
		client.WithCustomLogger(logger.NilLogger{}),
		client.WithNoMuxWriter(),
		client.WithMuxConnection(false),
		client.WithProtocol("std_tcp"),
		client.WithStackTrace(),
		client.WithAddress(*serverAddr),
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

func getAllInstance(ctx context.Context, proxy proxy.LittleRpcReflectionProxy, ot OutType, w io.Writer) {
	instances, err := proxy.AllInstance(ctx)
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

func getArgType(ctx context.Context, proxy proxy.LittleRpcReflectionProxy, service string, ot OutType, w io.Writer) {
	argType, err := proxy.MethodArgumentType(ctx, service)
	if err != nil {
		log.Fatalln(err)
	}
	switch ot {
	case Text:
		if argType == nil || len(argType) == 0 {
			return
		}
		for _, v := range argType {
			anyOutFromJson(v, Json, w)
		}
	default:
		anyOutFromJson(argType, ot, w)
	}
}

func getMethodTable(ctx context.Context, proxy proxy.LittleRpcReflectionProxy, sourceName string, ot OutType, w io.Writer) {
	tab, err := proxy.MethodTable(ctx, sourceName)
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

func getAllPacker(ctx context.Context, proxy proxy.LittleRpcReflectionProxy, ot OutType, w io.Writer) {
	packers, err := proxy.AllPacker(ctx)
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

func getAllCodec(ctx context.Context, proxy proxy.LittleRpcReflectionProxy, ot OutType, w io.Writer) {
	codecs, err := proxy.AllCodec(ctx)
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
		log.Fatalln("invalid output format")
	}
}

func callFunc(ctx context.Context, c Caller, service string, argsBytes [][]byte, ot OutType, w io.Writer) {
	args := make([]interface{}, len(argsBytes)+1)
	args[0] = ctx
	for k, rawArg := range argsBytes {
		err := json.Unmarshal(rawArg, &args[k+1])
		if err != nil {
			log.Fatalln(err)
		}
	}
	reps, err := c.RawCall(service, nil, args...)
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

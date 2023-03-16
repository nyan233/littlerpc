package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/sharedpool"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/nyan233/littlerpc/core/common/inters"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/common/metadata"
	"github.com/nyan233/littlerpc/core/common/msgparser"
	"github.com/nyan233/littlerpc/core/common/msgwriter"
	"github.com/nyan233/littlerpc/core/common/stream"
	transport2 "github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/common/utils/debug"
	metaDataUtil "github.com/nyan233/littlerpc/core/common/utils/metadata"
	"github.com/nyan233/littlerpc/core/container"
	lerror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/internal/pool"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
)

const (
	_Stop    int64 = 1 << 3
	_Start   int64 = 1 << 4
	_Restart int64 = 1 << 6
)

type connSourceDesc struct {
	Parser     msgparser.Parser
	Writer     msgwriter.Writer
	remoteAddr net.Addr
	localAddr  net.Addr
	cacheCtx   context.Context
	ctxManager *contextManager
}

type Server struct {
	//	Server 提供的所有服务, 一个服务即一个API
	//	比如: /littlerpc/internal/reflection.GetHandler --> func (l *LittleRpcReflection) GetHandler(source string) MethodTable
	services *container.RCUMap[string, *metadata.Process]
	// Server 提供的所有资源, 一个资源即一组服务的集合
	// 主要供reflection使用
	sources *container.RCUMap[string, *metadata.Source]
	// Server Engine
	server transport2.ServerEngine
	// 任务池
	taskPool pool.TaskPool[string]
	// logger
	logger logger.LLogger
	// 注册的插件的管理器
	pManager *pluginManager
	// Error Handler
	eHandle lerror.LErrors
	// Server Global Event
	ev *event
	// 初始信息
	opts []Option
	// 用于原子更新
	config atomic.Pointer[Config]
	pool   *sharedpool.SharedPool
}

func New(opts ...Option) *Server {
	server := new(Server)
	server.pool = sharedpool.NewSharedPool()
	applyConfig(server, opts)
	server.ev = newEvent()
	server.services = container.NewRCUMap[string, *metadata.Process](128)
	server.sources = container.NewRCUMap[string, *metadata.Source](128)
	// init reflection service
	err := server.RegisterClass(ReflectionSource, &LittleRpcReflection{server}, nil)
	if err != nil {
		panic(err)
	}
	return server
}

func applyConfig(server *Server, opts []Option) {
	sc := &Config{}
	WithDefaultServer()(sc)
	for _, v := range opts {
		v(sc)
	}
	server.config.Store(sc)
	if sc.Logger != nil {
		server.logger = sc.Logger
	} else {
		server.logger = logger.DefaultLogger
	}
	builder := transport2.Manager.GetServerEngine(sc.NetWork)(transport2.NetworkServerConfig{
		Addrs:     sc.Address,
		KeepAlive: sc.KeepAlive,
		TLSPubPem: nil,
	})
	eventD := builder.EventDriveInter()
	eventD.OnMessage(server.onMessage)
	eventD.OnClose(server.onClose)
	eventD.OnOpen(server.onOpen)
	if sc.RegisterMPOnRead {
		eventD.OnRead(server.onRead)
	}
	// server engine
	server.server = builder.Server()
	// init plugin manager
	server.pManager = newPluginManager(sc.Plugins)
	// init ErrorHandler
	server.eHandle = sc.ErrHandler
	// New TaskPool
	if sc.ExecPoolBuilder != nil {
		server.taskPool = sc.ExecPoolBuilder.Builder(
			sc.PoolBufferSize, sc.PoolMinSize, sc.PoolMaxSize, debug.ServerRecover(server.logger))
	} else {
		server.taskPool = pool.NewTaskPool[string](
			sc.PoolBufferSize, sc.PoolMinSize, sc.PoolMaxSize, debug.ServerRecover(server.logger))
	}
}

func (s *Server) RegisterClass(source string, i interface{}, custom map[string]metadata.ProcessOption) error {
	type Setup func(ptr unsafe.Pointer)
	type SetupMethod struct {
		typ  unsafe.Pointer
		data Setup
	}

	if i == nil {
		return errors.New("register elem is nil")
	}
	src := &metadata.Source{}
	src.InstanceType = reflect.TypeOf(i)
	value := reflect.ValueOf(i)
	// 检查类型的名字是否正确
	if name := reflect.Indirect(value).Type().Name(); name == "" {
		return errors.New("the typ name is not defined")
	} else if name != "" && source == "" {
		source = name
	}
	// 检查是否有与该类型绑定的方法
	if value.NumMethod() == 0 {
		return errors.New("no bind receiver method")
	}
	// init map
	src.ProcessSet = make(map[string]*metadata.Process, src.InstanceType.NumMethod())
	rsType, rsIndex := s.scanRpcServer(src.InstanceType)
	var setup Setup
	for i := 0; i < src.InstanceType.NumMethod(); i++ {
		method := src.InstanceType.Method(i)
		if !method.IsExported() {
			continue
		}
		// 检查是否重写了RpcServer.Setup方法
		if rsType != nil {
			switch method.Name {
			case "Setup":
				// setup = value.Method(i).Interface().(func())
				// 如果使用以上方法设置setup func的debug时就看不见真正的执行代码
				// reflect.Value.Interface()在遇到Func类型时会使用makeMethodValue构造一个满足匿名函数定义的函数
				// 返回的函数不包含接收器, 可当成普通函数来使用, 运行时调用methodValueCall
				// debug RegisterClass时就会看不见原始函数的代码
				// 使用以下方式则不会, 因为获得的是原始的函数类型
				setupFun := *(*SetupMethod)(unsafe.Pointer(&method.Func))
				setup = setupFun.data
				continue
			case "HijackProcess":
				continue
			}
		}
		option, ok := custom[method.Name]
		if !ok {
			s.registerProcess(src, method.Name, value.Method(i), nil)
		} else {
			s.registerProcess(src, method.Name, value.Method(i), &option)
		}
	}
	eValue := value
	for eValue.Kind() == reflect.Ptr {
		eValue = eValue.Elem()
	}
	if rsType != nil && eValue.Field(rsIndex).CanSet() {
		eValue.Field(rsIndex).Set(reflect.ValueOf(RpcServer{
			Prefix: source,
			s:      s,
		}))
	}
	s.sources.Store(source, src)
	kvs := make([]container.RCUMapElement[string, *metadata.Process], 0, len(src.ProcessSet))
	for k, v := range src.ProcessSet {
		kvs = append(kvs, container.RCUMapElement[string, *metadata.Process]{
			Key:   fmt.Sprintf("%s.%s", source, k),
			Value: v,
		})
	}
	s.services.StoreMulti(kvs)
	if setup != nil {
		// offset := eValue.Type().Field(rsIndex).Offset
		// setup(unsafe.Pointer(uintptr(value.UnsafePointer()) + offset))
		setup(value.UnsafePointer())
	}
	return nil
}

func (s *Server) registerProcess(src *metadata.Source, process string, processValue reflect.Value, option *metadata.ProcessOption) {
	processDesc := &metadata.Process{
		Value:  processValue,
		Option: s.config.Load().DefaultProcessOption,
	}
	src.ProcessSet[process] = processDesc
	if option != nil {
		processDesc.Option = *option
	}
	// 一个参数都没有的话则不需要进行那些根据输入参数来调整的选项
	if processValue.Type().NumIn() == 0 {
		return
	}
	for j := 0; j < processValue.Type().NumIn(); j++ {
		processDesc.ArgsType = append(processDesc.ArgsType, processValue.Type().In(j))
	}
	jOffset := metaDataUtil.IFContextOrStream(processDesc, processValue.Type())
	if !processDesc.Option.CompleteReUsage {
		goto asyncCheck
	}
	for j := 0 + jOffset; j < processValue.Type().NumIn(); j++ {
		typ := processValue.Type().In(j)
		if !typ.Implements(reflect.TypeOf((*inters.Reset)(nil)).Elem()) {
			processDesc.Option.CompleteReUsage = false
			goto asyncCheck
		}
	}
	processDesc.Pool = sync.Pool{
		New: func() interface{} {
			inputs := reflect2.FuncInputTypeListReturnValue(processDesc.ArgsType, 0, func(i int) bool {
				return false
			}, true)
			switch {
			case processDesc.SupportContext && processDesc.SupportStream:
				inputs[0] = reflect.ValueOf(context.Background())
				inputs[1] = reflect.ValueOf(stream.LStream(nil))
			}
			return inputs
		},
	}
asyncCheck:
	if processDesc.Option.SyncCall {
		// TODO nop
	}
	return
}

func (s *Server) scanRpcServer(typ reflect.Type) (reflect.Type, int) {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.Anonymous {
			continue
		}
		if field.Type.Kind() != reflect.Struct {
			continue
		}
		if field.Type == reflect.TypeOf(RpcServer{}) {
			return field.Type, i
		}
	}
	return nil, -1
}

func (s *Server) RegisterAnonymousFunc(service string, fn interface{}, option *metadata.ProcessOption) error {
	if service == "" {
		return errors.New("service name is empty")
	}
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return errors.New("register type not a function")
	}
	source, ok := s.sources.LoadOk(AnonymousSource)
	if !ok {
		return errors.New("source not found")
	}
	s.registerProcess(source, service, reflect.ValueOf(fn), option)
	process := source.ProcessSet[service]
	if process != nil {
		return errors.New("register failed process is empty")
	}
	s.services.Store(fmt.Sprintf("%s.%s", AnonymousSource, service), process)
	return nil
}

func (s *Server) Service() error {
	s.ev.Entry(int(_Start))
	return s.service()
}

func (s *Server) service() error {
	err := s.server.Start()
	if err != nil {
		return err
	}
	done, ack, ok := s.ev.Wait()
	if !ok {
		return errors.New("wait event failed")
	}
	select {
	case <-done:
		defer ack()
	}
	err = s.server.Stop()
	if err != nil {
		return err
	}
	err = s.taskPool.Stop()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	if !(s.ev.Complete(int(_Start)) || s.ev.Complete(int(_Restart))) {
		return errors.New("server in unknown state")
	}
	return nil
}

// Restart TODO: 实现重启Server
func (s *Server) Restart(override bool, opts ...Option) error {
	if err := s.Stop(); err != nil {
		return err
	}
	s.ev.Entry(int(_Restart))
	if !override {
		s.opts = append(s.opts, opts...)
	} else {
		s.opts = opts
	}
	applyConfig(s, s.opts)
	return s.service()
}

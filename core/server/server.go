package server

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/inters"
	"github.com/nyan233/littlerpc/core/common/logger"
	"github.com/nyan233/littlerpc/core/common/metadata"
	"github.com/nyan233/littlerpc/core/common/msgparser"
	"github.com/nyan233/littlerpc/core/common/msgwriter"
	transport2 "github.com/nyan233/littlerpc/core/common/transport"
	"github.com/nyan233/littlerpc/core/common/utils/debug"
	metaDataUtil "github.com/nyan233/littlerpc/core/common/utils/metadata"
	"github.com/nyan233/littlerpc/core/container"
	lerror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/internal/pool"
	reflect2 "github.com/nyan233/littlerpc/internal/reflect"
	"net"
	"reflect"
	"sync"
)

type connSourceDesc struct {
	Parser     msgparser.Parser
	Writer     msgwriter.Writer
	remoteAddr net.Addr
	localAddr  net.Addr
	ctxManager *contextManager
}

type Server struct {
	//	Server 提供的所有服务, 一个服务即一个API
	//	比如: /littlerpc/internal/reflection.GetHandler --> func (l *LittleRpcReflection) GetHandler(source string) MethodTable
	services container.RCUMap[string, *metadata.Process]
	// Server 提供的所有资源, 一个资源即一组服务的集合
	// 主要供reflection使用
	sources container.RCUMap[string, *metadata.Source]
	// Server Engine
	server transport2.ServerEngine
	// 任务池
	taskPool pool.TaskPool[string]
	// 管理的连接与其拥有的资源
	connsSourceDesc container.RWMutexMap[transport2.ConnAdapter, *connSourceDesc]
	// logger
	logger logger.LLogger
	// 注册的插件的管理器
	pManager *pluginManager
	// Error Handler
	eHandle lerror.LErrors
	config  *Config
}

func New(opts ...Option) *Server {
	server := &Server{}
	sc := &Config{}
	WithDefaultServer()(sc)
	for _, v := range opts {
		v(sc)
	}
	server.config = sc
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
	server.services = *container.NewRCUMap[string, *metadata.Process]()
	server.sources = *container.NewRCUMap[string, *metadata.Source]()
	// init reflection service
	err := server.RegisterClass(ReflectionSource, &LittleRpcReflection{server}, nil)
	if err != nil {
		panic(err)
	}
	return server
}

func (s *Server) RegisterClass(source string, i interface{}, custom map[string]metadata.ProcessOption) error {
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
	for i := 0; i < src.InstanceType.NumMethod(); i++ {
		method := src.InstanceType.Method(i)
		if !method.IsExported() {
			continue
		}
		option, ok := custom[method.Name]
		if !ok {
			s.registerProcess(src, method.Name, value.Method(i), nil)
		} else {
			s.registerProcess(src, method.Name, value.Method(i), &option)
		}
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
	return nil
}

func (s *Server) registerProcess(src *metadata.Source, process string, processValue reflect.Value, option *metadata.ProcessOption) {
	processDesc := &metadata.Process{
		Value:  processValue,
		Option: new(metadata.ProcessOption),
	}
	src.ProcessSet[process] = processDesc
	if option != nil {
		processDesc.Option = option
	}
	// 一个参数都没有的话则不需要进行那些根据输入参数来调整的选项
	if processValue.Type().NumIn() == 0 {
		return
	}
	jOffset := metaDataUtil.IFContextOrStream(processDesc, processValue.Type())
	if !processDesc.Option.CompleteReUsage {
		goto asyncCheck
	}
	for j := 0 + jOffset; j < processValue.Type().NumIn(); j++ {
		if !processValue.Type().In(j).Implements(reflect.TypeOf((*inters.Reset)(nil)).Elem()) {
			processDesc.Option.CompleteReUsage = false
			goto asyncCheck
		}
	}
	processDesc.Pool = sync.Pool{
		New: func() interface{} {
			tmp := make([]reflect.Value, jOffset, 8)
			inputs := reflect2.FuncInputTypeList(processDesc.Value, jOffset, false, func(i int) bool {
				return false
			})
			for _, v := range inputs {
				tmp = append(tmp, reflect.ValueOf(v))
			}
			return tmp
		},
	}
asyncCheck:
	if processDesc.Option.SyncCall {
		// TODO nop
	}
	return
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

func (s *Server) Start() error {
	return s.server.Start()
}

func (s *Server) Stop() error {
	err := s.taskPool.Stop()
	if err != nil {
		return err
	}
	return s.server.Stop()
}

// Restart TODO: 实现重启Server
func (s *Server) Restart(opts ...Option) error {
	return errors.New("restart not implemented")
}

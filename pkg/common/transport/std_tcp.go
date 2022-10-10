package transport

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

const (
	StdTCPClient int = iota
	StdTCPServer
)

type StdNetTcpEngine struct {
	mu sync.Mutex
	// 指示是客户端模式还是服务器
	mode      int
	onOpen    func(conn ConnAdapter)
	onMessage func(conn ConnAdapter, data []byte)
	onClose   func(conn ConnAdapter, err error)
	addrs     []string
	listeners []net.Listener
	readBuf   sync.Pool
	closed    int32
}

func NewStdTcpServerEngine(config NetworkServerConfig) ServerEngineBuilder {
	return &StdNetTcpEngine{
		listeners: make([]net.Listener, len(config.Addrs)),
		addrs:     config.Addrs,
		mode:      StdTCPServer,
		readBuf: sync.Pool{
			New: func() interface{} {
				tmp := make([]byte, ReadBufferSize)
				return &tmp
			},
		},
		onOpen:    func(conn ConnAdapter) {},
		onMessage: func(conn ConnAdapter, data []byte) {},
		onClose:   func(conn ConnAdapter, err error) {},
	}
}

func NewStdTcpClientEngine() ClientEngineBuilder {
	return &StdNetTcpEngine{
		mode: StdTCPClient,
		readBuf: sync.Pool{
			New: func() interface{} {
				tmp := make([]byte, ReadBufferSize)
				return &tmp
			},
		},
		onOpen:    func(conn ConnAdapter) {},
		onMessage: func(conn ConnAdapter, data []byte) {},
		onClose:   func(conn ConnAdapter, err error) {},
	}
}

func (s *StdNetTcpEngine) NewConn(config NetworkClientConfig) (ConnAdapter, error) {
	conn, err := net.Dial("tcp", config.ServerAddr)
	if err != nil {
		return nil, err
	}
	defer s.connService(conn)
	return conn, nil
}

func (s *StdNetTcpEngine) Server() ServerEngine {
	return s
}

func (s *StdNetTcpEngine) Client() ClientEngine {
	return s
}

func (s *StdNetTcpEngine) EventDriveInter() EventDriveInter {
	return s
}

func (s *StdNetTcpEngine) Start() error {
	if atomic.LoadInt32(&s.closed) == 1 {
		return errors.New("wsEngine already closed")
	}
	if s.mode == StdTCPClient {
		return nil
	}
	var wg sync.WaitGroup
	wg.Add(len(s.listeners))
	for k, v := range s.addrs {
		lIndex := k
		addr := v
		go func() {
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				panic(err)
			}
			wg.Done()
			s.mu.Lock()
			s.listeners[lIndex] = listener
			s.mu.Unlock()
			for {
				conn, err := listener.Accept()
				if err != nil {
					s.onClose(conn, err)
					break
				}
				s.connService(conn)
			}
		}()
	}
	wg.Wait()
	return nil
}

func (s *StdNetTcpEngine) connService(conn net.Conn) {
	s.onOpen(conn)
	go func() {
		for {
			if atomic.LoadInt32(&s.closed) == 1 {
				s.onClose(conn, errors.New("eventLoop already closed"))
				_ = conn.Close()
				break
			}
			readBuf := s.readBuf.Get().(*[]byte)
			readN, err := conn.Read(*readBuf)
			if err != nil {
				s.readBuf.Put(readBuf)
				s.onClose(conn, err)
				_ = conn.Close()
				break
			}
			s.onMessage(conn, (*readBuf)[:readN])
			s.readBuf.Put(readBuf)
		}
	}()
}

func (s *StdNetTcpEngine) Stop() error {
	if !atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		return errors.New("wsEngine already closed")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.listeners {
		_ = v.Close()
	}
	return nil
}

func (s *StdNetTcpEngine) OnMessage(f func(conn ConnAdapter, data []byte)) {
	s.onMessage = f
}

func (s *StdNetTcpEngine) OnOpen(f func(conn ConnAdapter)) {
	s.onOpen = f
}

func (s *StdNetTcpEngine) OnClose(f func(conn ConnAdapter, err error)) {
	s.onClose = f
}

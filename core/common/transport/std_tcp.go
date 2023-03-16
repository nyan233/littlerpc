package transport

import (
	"errors"
	"github.com/nyan233/littlerpc/core/common/logger"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
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
	onRead    func(conn ConnAdapter)
	onMessage func(conn ConnAdapter, data []byte)
	onClose   func(conn ConnAdapter, err error)
	addrs     []string
	listeners []net.Listener
	readBuf   sync.Pool
	closed    int32
}

func NewStdTcpServer(config NetworkServerConfig) ServerBuilder {
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

func NewStdTcpClient() ClientBuilder {
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
	return s.connService(conn), nil
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
					logger.DefaultLogger.Warn("std-tcp engine accept conn failed, err = %v", err)
					break
				}
				s.connService(conn)
			}
		}()
	}
	wg.Wait()
	return nil
}

func (s *StdNetTcpEngine) connService(conn net.Conn) *nioConn {
	nc := &nioConn{Conn: conn}
	s.onOpen(nc)
	go func() {
		var (
			buf = make([]byte, 0)
		)
		for {
			if atomic.LoadInt32(&s.closed) == 1 {
				s.onClose(nc, errors.New("eventLoop already closed"))
				_ = nc.Close()
				break
			}
			if s.onRead != nil {
				_, err := conn.Read(buf)
				if err != nil {
					s.onClose(nc, err)
					_ = nc.Close()
					break
				}
				s.onRead(nc)
				continue
			}
			readBuf := s.readBuf.Get().(*[]byte)
			readN, err := conn.Read(*readBuf)
			if err != nil {
				s.readBuf.Put(readBuf)
				s.onClose(nc, err)
				_ = nc.Close()
				break
			}
			s.onMessage(nc, (*readBuf)[:readN])
			s.readBuf.Put(readBuf)
		}
	}()
	return nc
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

func (s *StdNetTcpEngine) OnRead(f func(conn ConnAdapter)) {
	s.onRead = f
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

type nioConn struct {
	net.Conn
	source atomic.Value
}

// 保证OnRead只调用一次Read
func (c *nioConn) Read(p []byte) (n int, err error) {
	readN, err := c.Conn.Read(p)
	if err != nil {
		return readN, err
	}
	return readN, syscall.EWOULDBLOCK
}

func (c *nioConn) SetSource(s interface{}) {
	c.source.Store(s)
}

func (c *nioConn) Source() interface{} {
	return c.source.Load()
}

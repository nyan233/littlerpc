package transport

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

type StdTcpOption struct {
	Network           string
	MaxReadBufferSize int
	Addrs             []string
}

type StdTcpServer struct {
	mu        sync.Mutex
	onOpen    func(conn ConnAdapter)
	onMessage func(conn ConnAdapter, data []byte)
	onClose   func(conn ConnAdapter, err error)
	onError   func(err error)
	option    *StdTcpOption
	listeners []net.Listener
	readBuf   sync.Pool
	closed    int32
}

func NewStdTcpTransServer(opt *StdTcpOption) ServerTransportBuilder {
	return &StdTcpServer{
		option:    opt,
		listeners: make([]net.Listener, len(opt.Addrs)),
		readBuf: sync.Pool{
			New: func() interface{} {
				tmp := make([]byte, opt.MaxReadBufferSize)
				return &tmp
			},
		},
	}
}

func (s *StdTcpServer) Instance() ServerTransport {
	return s
}

func (s *StdTcpServer) Start() error {
	if atomic.LoadInt32(&s.closed) == 1 {
		return errors.New("server already closed")
	}
	var wg sync.WaitGroup
	wg.Add(len(s.option.Addrs))
	for k, v := range s.option.Addrs {
		lIndex := k
		addr := v
		go func() {
			listener, err := net.Listen(s.option.Network, addr)
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
					s.onError(err)
					break
				}
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
							s.onError(err)
							s.onClose(conn, err)
							_ = conn.Close()
							break
						}
						s.onMessage(conn, (*readBuf)[:readN])
						s.readBuf.Put(readBuf)
					}
				}()
			}
		}()
	}
	wg.Wait()
	return nil
}

func (s *StdTcpServer) Stop() error {
	if !atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		return errors.New("server already closed")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.listeners {
		_ = v.Close()
	}
	return nil
}

func (s *StdTcpServer) SetOnMessage(f func(conn ConnAdapter, data []byte)) {
	s.onMessage = f
}

func (s *StdTcpServer) SetOnClose(f func(conn ConnAdapter, err error)) {
	s.onClose = f
}

func (s *StdTcpServer) SetOnOpen(f func(conn ConnAdapter)) {
	s.onOpen = f
}

func (s *StdTcpServer) SetOnErr(f func(err error)) {
	s.onError = f
}

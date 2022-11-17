package selector

import (
	"github.com/nyan233/littlerpc/pkg/common/transport"
)

type arrayBaseConnSelector struct {
	newConn      func() (transport.ConnAdapter, error)
	connPoolSize int
	conns        []transport.ConnAdapter
}

func newArrayBaseConnSelector(poolSize int, newConn func() (transport.ConnAdapter, error)) ConnSelector {
	return &arrayBaseConnSelector{
		newConn:      newConn,
		connPoolSize: poolSize,
	}
}

func (r *arrayBaseConnSelector) Take() (transport.ConnAdapter, error) {
	return nil, nil
}

func (r *arrayBaseConnSelector) Acquire(conn transport.ConnAdapter) {
	r.conns = append(r.conns, conn)
}

func (r *arrayBaseConnSelector) Release(conn transport.ConnAdapter) {
	if r.conns == nil || len(r.conns) == 0 {
		return
	}
	for k, v := range r.conns {
		if v == conn {
			r.conns[k] = nil
			for i := k; i < len(r.conns); i++ {
				if k == len(r.conns)-1 {
					continue
				}
				r.conns[k] = r.conns[k+1]
			}
		}
	}
}

func (r *arrayBaseConnSelector) Close() int {
	var closeCount int
	for _, v := range r.conns {
		err := v.Close()
		if err == nil {
			closeCount++
		}
	}
	return closeCount
}

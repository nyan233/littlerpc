package selector

import (
	"github.com/nyan233/littlerpc/pkg/common/transport"
	random2 "github.com/nyan233/littlerpc/pkg/utils/random"
)

type randomConnSelector struct {
	*arrayBaseConnSelector
}

func newRandomConnSelector(poolSize int, newConn func() (transport.ConnAdapter, error)) ConnSelector {
	return &randomConnSelector{
		newArrayBaseConnSelector(poolSize, newConn),
	}
}

func (r *randomConnSelector) Take() (transport.ConnAdapter, error) {
	if r.conns == nil || len(r.conns) == 0 {
		for i := 0; i < r.connPoolSize; i++ {
			conn, err := r.newConn()
			if err != nil {
				return nil, err
			}
			r.conns = append(r.conns, conn)
		}
	}
	return r.conns[int(random2.FastRand())%len(r.conns)], nil
}

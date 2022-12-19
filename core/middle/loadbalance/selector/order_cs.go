package selector

import (
	"github.com/nyan233/littlerpc/core/common/transport"
)

type orderConnSelector struct {
	*arrayBaseConnSelector
	count int
}

func newOrderConnSelector(poolSize int, newConn func() (transport.ConnAdapter, error)) ConnSelector {
	return &orderConnSelector{
		newArrayBaseConnSelector(poolSize, newConn),
		0,
	}
}

func (o *orderConnSelector) Take() (transport.ConnAdapter, error) {
	if o.conns == nil || len(o.conns) == 0 {
		for i := 0; i < o.connPoolSize; i++ {
			conn, err := o.newConn()
			if err != nil {
				return nil, err
			}
			o.conns = append(o.conns, conn)
		}
	}
	o.count++
	return o.conns[o.count%len(o.conns)], nil
}

package stream

import "github.com/nyan233/littlerpc/common/transport"

type LStream interface {
	RecvMsg(data interface{}) error
	SendMsg(data interface{}) error
	Close(flag int) error
}

type LStreamFactory interface {
	getLStreamConn() (transport.ConnAdapter, error)
}

func NewLStream(factory LStreamFactory) (LStream, error) {
	return nil, nil
}

package server

import (
	"context"
	"errors"
	"github.com/nyan233/littlerpc/pkg/common/transport"
	"github.com/nyan233/littlerpc/pkg/container"
)

type contextManager struct {
	manager container.RWMutexMap[transport.ConnAdapter, *container.MutexMap[uint64, context.CancelFunc]]
}

func (manager *contextManager) RegisterConnection(conn transport.ConnAdapter) {
	_, ok := manager.manager.LoadOk(conn)
	if ok {
		return
	}
	manager.manager.Store(conn, &container.MutexMap[uint64, context.CancelFunc]{})
}

func (manager *contextManager) DeleteConnection(conn transport.ConnAdapter) {
	manager.manager.Store(conn, nil)
	manager.manager.Delete(conn)
}

func (manager *contextManager) RegisterContextCancel(conn transport.ConnAdapter, contextId uint64, cancel context.CancelFunc) error {
	ctxCollect, ok := manager.manager.LoadOk(conn)
	if !ok {
		return errors.New("context collection not found")
	}
	ctxCollect.Store(contextId, cancel)
	return nil
}

func (manager *contextManager) CancelContext(conn transport.ConnAdapter, contextId uint64) error {
	ctxCollect, ok := manager.manager.LoadOk(conn)
	if !ok {
		return errors.New("context collection not found")
	}
	cancel, ok := ctxCollect.LoadOk(contextId)
	if !ok {
		return errors.New("context cancel func not found")
	}
	cancel()
	ctxCollect.Store(contextId, nil)
	ctxCollect.Delete(contextId)
	return nil
}

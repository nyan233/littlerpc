package server

import (
	"context"
	"errors"
	"github.com/nyan233/littlerpc/core/container"
)

type contextManager struct {
	ctxSet container.MutexMap[uint64, context.CancelFunc]
}

func newContextManager() *contextManager {
	return &contextManager{
		ctxSet: container.MutexMap[uint64, context.CancelFunc]{},
	}
}

func (cm *contextManager) RegisterContextCancel(contextId uint64, cancel context.CancelFunc) error {
	cm.ctxSet.Store(contextId, cancel)
	return nil
}

func (cm *contextManager) CancelContext(contextId uint64) error {
	cancel, ok := cm.ctxSet.LoadOk(contextId)
	if !ok {
		return errors.New("context cancel func not found")
	}
	cancel()
	cm.ctxSet.Store(contextId, nil)
	cm.ctxSet.Delete(contextId)
	return nil
}

func (cm *contextManager) CancelAll() {
	cm.ctxSet.Range(func(ctxId uint64, cancel context.CancelFunc) bool {
		cancel()
		return true
	})
}

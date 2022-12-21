package server

import (
	"context"
	"errors"
	"sync"
)

type elem struct {
	Context context.Context
	Cancel  context.CancelFunc
}

type contextManager struct {
	mu     sync.Mutex
	ctxSet map[uint64]elem
}

func newContextManager() *contextManager {
	return &contextManager{
		ctxSet: make(map[uint64]elem),
	}
}

func (cm *contextManager) RegisterContextCancel(contextId uint64) (context.Context, context.CancelFunc) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	val, ok := cm.ctxSet[contextId]
	if ok {
		return val.Context, val.Cancel
	}
	ctx, cancel := context.WithCancel(context.Background())
	cm.ctxSet[contextId] = elem{
		Context: ctx,
		Cancel:  cancel,
	}
	return ctx, cancel
}

func (cm *contextManager) CancelContext(contextId uint64) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	val, ok := cm.ctxSet[contextId]
	if !ok {
		return errors.New("context cancel func not found")
	}
	val.Cancel()
	delete(cm.ctxSet, contextId)
	return nil
}

func (cm *contextManager) CancelAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, v := range cm.ctxSet {
		v.Cancel()
	}
}

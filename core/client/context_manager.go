package client

import (
	"context"
	"sync"
)

type contextManager struct {
	lock     sync.Mutex
	ctxIdSet map[context.Context]uint64
}

func newContextManager() *contextManager {
	return &contextManager{
		ctxIdSet: make(map[context.Context]uint64, 16),
	}
}

// Register 返回有效的Id
func (cm *contextManager) Register(ctx context.Context, id uint64) (_ uint64) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cid, ok := cm.ctxIdSet[ctx]
	if ok {
		return cid
	}
	cm.ctxIdSet[ctx] = id
	return id
}

// Unregister 返回有效的Id
func (cm *contextManager) Unregister(ctx context.Context) (_ uint64) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cid, ok := cm.ctxIdSet[ctx]
	if !ok {
		return 0
	}
	delete(cm.ctxIdSet, ctx)
	return cid
}

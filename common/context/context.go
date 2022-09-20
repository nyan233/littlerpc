package context

import (
	"context"
	"time"
)

type CancelFunc = context.CancelFunc

// LContext 与标准的context.context差别不大, 但是这个使用这个context
// 可以获得性能上的优势
type LContext struct {
	context.Context
}

func WithCancel(parent context.Context) (context.Context, CancelFunc) {
	return context.WithCancel(parent)
}

func WithDeadline(parent context.Context, deadline time.Time) (context.Context, CancelFunc) {
	return context.WithDeadline(parent, deadline)
}

func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

func WithValue(parent context.Context, key, value any) context.Context {
	return context.WithValue(parent, key, value)
}

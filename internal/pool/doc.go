// Package pool littlerpc自带的goroutine池
package pool

type Hash interface {
	string | []byte | int64 | uint64
}

type TaskPool[Key Hash] interface {
	Push(Key, func()) error
	Stop() error
	// LiveSize 存活的goroutine数量
	LiveSize() int
	// BufSize 缓冲区中存在的任务数量
	BufSize() int
	// ExecuteSuccess 任务池执行成功的任务数量
	ExecuteSuccess() int
	// ExecuteError 任务池执行失败的任务数量
	ExecuteError() int
}

type TaskPoolBuilder[Key Hash] interface {
	Builder(bufSize, minSize, maxSize int32, rf RecoverFunc) TaskPool[Key]
}

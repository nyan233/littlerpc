package export

import "github.com/nyan233/littlerpc/internal/pool"

// 用于导出internal包中一些需要暴露的接口

type TaskPool interface {
	pool.TaskPool
}

type TaskPoolBuilder interface {
	pool.TaskPoolBuilder
}

type Reset interface {
	Reset()
}

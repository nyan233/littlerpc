package export

import "github.com/nyan233/littlerpc/internal/pool"

// 用于导出internal包中一些需要暴露的接口

type TaskPool[Key pool.Hash] interface {
	pool.TaskPool[Key]
}

type TaskPoolBuilder[Key pool.Hash] interface {
	pool.TaskPoolBuilder[Key]
}

type Reset interface {
	Reset()
}

package server

import (
	"github.com/nyan233/littlerpc/internal/pool"
)

type poolBuilder[Key pool.Hash] func(bufSize, minSize, maxSize int32, rf pool.RecoverFunc) pool.TaskPool[Key]

func (p poolBuilder[Key]) Builder(bufSize, minSize, maxSize int32, rf pool.RecoverFunc) pool.TaskPool[Key] {
	return p(bufSize, minSize, maxSize, rf)
}

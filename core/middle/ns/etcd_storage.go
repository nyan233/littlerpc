package ns

import (
	"github.com/nyan233/littlerpc/core/container"
)

type etcdStorage struct {
}

func NewEtcdStorage(cfg StorageConfig) Storage {
	return &etcdStorage{}
}

func (e *etcdStorage) GetNodeList(key string) (int, container.Slice[Node], error) {
	//TODO implement me
	panic("implement me")
}

func (e *etcdStorage) SetUpdateCallback(f func(key string, version int, nodeList container.Slice[Node])) {
	//TODO implement me
	panic("implement me")
}

func (e *etcdStorage) SetNodeList(key string, version int, nodeList container.Slice[Node]) error {
	//TODO implement me
	panic("implement me")
}

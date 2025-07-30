package ns

type etcdStorage struct {
}

func NewEtcdStorage(cfg StorageConfig) Storage {
	return &etcdStorage{}
}

func (e *etcdStorage) Start() error {
	//TODO implement me
	panic("implement me")
}

func (e *etcdStorage) GetNodeList(key string) (int, []Node, error) {
	//TODO implement me
	panic("implement me")
}

func (e *etcdStorage) SetUpdateCallback(f func(key string, version int, nodeList []Node)) {
	//TODO implement me
	panic("implement me")
}

func (e *etcdStorage) SetNodeList(key string, version int, nodeList []Node) error {
	//TODO implement me
	panic("implement me")
}

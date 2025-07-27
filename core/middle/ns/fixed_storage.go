package ns

import (
	"net"
	"strconv"
)

type fixedStorage struct {
	nodeList []Node
	updateCb func(key string, version int, nodeList []Node)
}

func NewFixedStorage(addrList []string) Storage {
	s := new(fixedStorage)
	for _, addr := range addrList {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			panic(err)
		}
		portUint, err := strconv.ParseUint(port, 10, 64)
		if err != nil {
			panic(err)
		}
		s.nodeList = append(s.nodeList, Node{
			Ip:       net.ParseIP(host),
			Port:     uint32(portUint),
			Addr:     addr,
			Priority: 1,
		})
	}
	return s
}

func (f *fixedStorage) Start() error {
	f.updateCb("", 0, f.nodeList)
	return nil
}

func (f *fixedStorage) GetNodeList(key string) (int, []Node, error) {
	return 0, f.nodeList, nil
}

func (f *fixedStorage) SetUpdateCallback(f2 func(key string, version int, nodeList []Node)) {
	f.updateCb = f2
}

func (f *fixedStorage) SetNodeList(key string, version int, nodeList []Node) error {
	return nil
}

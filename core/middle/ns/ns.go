package ns

import (
	"errors"
	"fmt"
	"github.com/lafikl/consistent"
	"github.com/nyan233/littlerpc/core/container"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"github.com/nyan233/littlerpc/core/utils/hash"
	"github.com/nyan233/littlerpc/core/utils/random"
	"slices"
	"strings"
	"sync"
)

/*
	littlerpc name server api
*/

const (
	SchemeRandom         = "random"
	SchemeHash           = "hash"
	SchemeConsistentHash = "consistent-hash"
)

type StorageConfig struct {
	// name or ipaddr
	Addr string
}

type Storage interface {
	Start() error
	GetNodeList(key string) (int, []Node, error)
	SetUpdateCallback(func(key string, version int, nodeList []Node))
	SetNodeList(key string, version int, nodeList []Node) error
}

type Node struct {
	Ip   []byte
	Port uint32
	Addr string
	// 1~10
	Priority int
}

type Config struct {
	Storage Storage
	Scheme  string
}

type nsCacheVal struct {
	NodeList   []Node
	CHNodeList *consistent.Consistent
}

type NameServer struct {
	cfg      Config
	mu       sync.Mutex
	cache    *container.SyncMap118[string, nsCacheVal]
	storeApi Storage
}

func NewNameServer(cfg Config) *NameServer {
	ns := &NameServer{
		cfg:      cfg,
		cache:    new(container.SyncMap118[string, nsCacheVal]),
		storeApi: cfg.Storage,
	}
	ns.storeApi.SetUpdateCallback(ns.updateCallback)
	return ns
}

func (ns *NameServer) Close() error {
	return nil
}

func (ns *NameServer) Start() error {
	return ns.storeApi.Start()
}

func (ns *NameServer) genCacheVal(nodeList []Node) nsCacheVal {
	cv := nsCacheVal{
		NodeList:   nodeList,
		CHNodeList: consistent.New(),
	}
	for _, node := range nodeList {
		cv.CHNodeList.Add(node.Addr)
	}
	return cv
}

func (ns *NameServer) updateCallback(key string, version int, nodeList []Node) {
	ns.cache.Store(key, ns.genCacheVal(nodeList))
}

// GetNode 首次适应
func (ns *NameServer) GetNode(path string) (Node, error) {
	s := strings.Split(path, ".")
	if len(s) != 2 {
		panic(fmt.Sprintf("unknown path format : %s", path))
	}
	class := s[0]
	method := s[1]
	cv, ok := ns.cache.LoadOk(class)
	if !ok {
		ns.mu.Lock()
		defer ns.mu.Unlock()
		cv, ok = ns.cache.LoadOk(class)
		if ok {
			goto success
		}
		var (
			nodeList []Node
			err      error
		)
		_, nodeList, err = ns.storeApi.GetNodeList(class)
		if err != nil {
			return Node{}, err
		}
		cv = ns.genCacheVal(nodeList)
		ns.cache.Store(class, cv)
	}
success:
	nNode := len(cv.NodeList)
	if nNode <= 0 {
		return Node{}, errors.New("node list is empty")
	} else if nNode == 1 {
		return cv.NodeList[0], nil
	} else {
		switch ns.cfg.Scheme {
		case SchemeRandom:
			return cv.NodeList[random.FastRand()%uint32(nNode)], nil
		case SchemeHash:
			const HashSeed = 1<<12 + 1
			hashCode := hash.Murmurhash3Onx8632(convert.StringToBytes(method), HashSeed)
			return cv.NodeList[hashCode%uint32(nNode)], nil
		case SchemeConsistentHash:
			addr, err := cv.CHNodeList.Get(method)
			if err != nil {
				return Node{}, err
			}
			idx := slices.IndexFunc(cv.NodeList, func(node Node) bool {
				return node.Addr == addr
			})
			return cv.NodeList[idx], nil
		default:
			panic(fmt.Sprintf("unknown scheme : %s", ns.cfg.Scheme))
		}
	}
}

package loadbalance

import (
	"io"
	"net/http"
	"os"
	"strings"
)

func DefaultFileResolver(path string) ResolverFunc {
	return func() ([]RpcNode, error) {
		fileData, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		nodeAddrs := strings.Split(string(fileData), "\n")
		nodes := make([]RpcNode, 0, len(nodeAddrs))
		for _, nodeAddr := range nodeAddrs {
			nodes = append(nodes, RpcNode{Address: nodeAddr})
		}
		return nodes, nil
	}
}

func DefaultHttpResolver(url string) ResolverFunc {
	return func() ([]RpcNode, error) {
		response, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		nodeAddrs := strings.Split(string(bytes), "\n")
		nodes := make([]RpcNode, 0, len(nodeAddrs))
		for _, nodeAddr := range nodeAddrs {
			nodes = append(nodes, RpcNode{Address: nodeAddr})
		}
		return nodes, nil
	}
}

func DefaultLiveResolver(splitAddr string) ResolverFunc {
	return func() ([]RpcNode, error) {
		nodeAddrs := strings.Split(splitAddr, ";")
		nodes := make([]RpcNode, 0, len(nodeAddrs))
		for _, nodeAddr := range nodeAddrs {
			nodes = append(nodes, RpcNode{Address: nodeAddr})
		}
		return nodes, nil
	}
}

package main

import (
	"fmt"
	"github.com/nyan233/littlerpc/core/client"
	"github.com/nyan233/littlerpc/core/server"
	"io/ioutil"
	"os"
)

type FileServer struct {
	fileMap map[string][]byte
}

func New() *FileServer {
	return &FileServer{fileMap: make(map[string][]byte)}
}

func (fs *FileServer) SendFile(path string, data []byte) error {
	fs.fileMap[path] = data
	return nil
}

func (fs *FileServer) GetFile(path string) ([]byte, bool, error) {
	bytes, ok := fs.fileMap[path]
	return bytes, ok, nil
}

func (fs *FileServer) OpenSysFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(file)
}

func main() {
	server := server.New(server.WithAddressServer(":1234"))
	_ = server.RegisterClass("", New(), nil)
	go server.Service()
	client, err := client.New(client.WithAddress(":1234"))
	if err != nil {
		panic(err)
	}
	proxy := NewFileServer(client)
	fileBytes, err := proxy.OpenSysFile("./main.go")
	if err != nil {
		panic(err)
	}
	err = proxy.SendFile("main.go", fileBytes)
	if err != nil {
		panic(err)
	}
	fileBytes, ok, _ := proxy.GetFile("main.go")
	if !ok {
		panic("no such file")
	}
	fmt.Println(string(fileBytes))
}

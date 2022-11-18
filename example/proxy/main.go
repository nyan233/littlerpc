package main

import (
	"fmt"
	"github.com/nyan233/littlerpc/client"
	"github.com/nyan233/littlerpc/server"
	"io/ioutil"
	"os"
)

type FileServer struct {
	fileMap map[string][]byte
}

func NewFileServer() *FileServer {
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
	_ = server.RegisterClass("", NewFileServer(), nil)
	err := server.Start()
	if err != nil {
		panic(err)
	}
	client, err := client.New(client.WithAddress(":1234"))
	if err != nil {
		panic(err)
	}
	proxy := NewFileServerProxy(client)
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

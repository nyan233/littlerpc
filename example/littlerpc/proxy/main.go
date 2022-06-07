package main

import (
	"fmt"
	"github.com/nyan233/littlerpc"
	"io/ioutil"
	"os"
)

type FileServer struct {
	fileMap map[string][]byte
}

func NewFileServer() *FileServer {
	return &FileServer{fileMap: make(map[string][]byte)}
}

func (fs *FileServer) SendFile(path string, data []byte) {
	fs.fileMap[path] = data
}

func (fs *FileServer) GetFile(path string) ([]byte, bool) {
	bytes, ok := fs.fileMap[path]
	return bytes, ok
}

func (fs *FileServer) OpenSysFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(file)
}

func main() {
	server := littlerpc.NewServer(littlerpc.WithAddressServer(":1234"))
	_ = server.Elem(NewFileServer())
	err := server.Start()
	if err != nil {
		panic(err)
	}
	client := littlerpc.NewClient(littlerpc.WithAddressClient(":1234"))
	proxy := NewFileServerProxy(client)
	fileBytes, err := proxy.OpenSysFile("./main.go")
	if err != nil {
		panic(err)
	}
	proxy.SendFile("main.go", fileBytes)
	fileBytes, ok := proxy.GetFile("main.go")
	if !ok {
		panic("no such file")
	}
	fmt.Println(string(fileBytes))
}

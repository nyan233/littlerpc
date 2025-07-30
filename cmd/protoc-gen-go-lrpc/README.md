## Install
```shell
go install github.com/nyan233/littlerpc/cmd/protoc-gen-go-lrpc@latest
```
## Run
```shell
protoc --go_out=. --go-lrpc_out=. --proto_path=. namsserver.proto
```
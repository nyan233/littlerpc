package main

import (
	"google.golang.org/protobuf/proto"
)

type ProtoBufCodec struct{}

func (p ProtoBufCodec) Scheme() string {
	return "protobuf"
}

func (p ProtoBufCodec) Marshal(i interface{}) ([]byte, error) {
	return proto.Marshal(i.(proto.Message))
}

func (p ProtoBufCodec) Unmarshal(data []byte, i interface{}) error {
	return proto.Unmarshal(data, i.(proto.Message))
}

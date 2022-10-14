package main

import (
	"errors"
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

func (p ProtoBufCodec) UnmarshalError(data []byte, v interface{}) error {
	switch v.(type) {
	case *error:
		// 这种情况表示nil
		if len(data) == 1 && data[0] == 0 {
			return nil
		}
		str := string(data)
		*(v.(*error)) = errors.New(str)
		return nil
	default:
		return errors.New("no support type")
	}
}

func (p ProtoBufCodec) MarshalError(v interface{}) ([]byte, error) {
	if e, ok := v.(error); ok {
		return []byte(e.Error()), nil
	} else if v == nil {
		return []byte{0}, nil
	} else {
		return nil, errors.New("no error type")
	}
}

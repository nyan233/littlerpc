package protocol

import (
	"encoding/json"
)

type Codec interface {
	Scheme() string
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte,v interface{}) error
}

var (
	codecCollection = make(map[string]Codec)
)

// RegisterCodec 该调用不是线程安全的
func RegisterCodec(c Codec) {
	codecCollection[c.Scheme()] = c
}

// GetCodec 该调用不是线程安全的
func GetCodec(scheme string) Codec {
	return codecCollection[scheme]
}

type JsonCodec struct {}

func (j JsonCodec) Scheme() string {
	return "json"
}

func (j JsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j JsonCodec) Unmarshal(data []byte,v interface{}) error {
	return json.Unmarshal(data,v)
}

func init() {
	RegisterCodec(new(JsonCodec))
}
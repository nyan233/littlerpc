package codec

import (
	"encoding/json"
)

type Codec interface {
	Scheme() string
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

var (
	codecCollection = make(map[string]Codec, 8)
)

// Register 注册一个Code, 按照定义的scheme获取Codec
// 从v0.4.0版本开始LittleRpc在Codec中使用普通map来管理
func Register(c Codec) {
	if c == nil {
		panic("codec is nil")
	}
	if c.Scheme() == "" {
		panic("codec scheme is empty")
	}
	codecCollection[c.Scheme()] = c
}

// Get 根据Scheme获取Codec, 如果Codec不存在, 那么返回的Codec == nil
// 从v0.4.0版本开始LittleRpc在Codec中使用普通map来管理
func Get(scheme string) Codec {
	return codecCollection[scheme]
}

type Json struct{}

func (j Json) Scheme() string {
	return "json"
}

func (j Json) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j Json) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func init() {
	Register(new(Json))
}

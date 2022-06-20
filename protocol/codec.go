package protocol

import (
	"encoding/json"
	"errors"
)

type Codec interface {
	Scheme() string
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	// UnmarshalError 负责反序列化error类型
	UnmarshalError(data []byte, v interface{}) error
	// MarshalError 负责序列化error类型
	MarshalError(v interface{}) ([]byte, error)
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

type JsonCodec struct{}

func (j JsonCodec) Scheme() string {
	return "json"
}

func (j JsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j JsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (j JsonCodec) UnmarshalError(data []byte, v interface{}) error {
	if len(data) == 1 && data[0] == 0 {
		return nil
	}
	switch v.(type) {
	case *Error:
		return json.Unmarshal(data, v)
	case *error:
		var str string
		err := json.Unmarshal(data, &str)
		if err != nil {
			return err
		}
		*(v.(*error)) = errors.New(str)
		return nil
	default:
		return errors.New("type not error")
	}
}

func (j JsonCodec) MarshalError(v interface{}) ([]byte, error) {
	switch v.(type) {
	case *Error:
		return json.Marshal(v)
	case error:
		var str = v.(error).Error()
		return json.Marshal(&str)
	case nil:
		var i int64
		return json.Marshal(&i)
	default:
		return nil, errors.New("type not error")
	}
}

func init() {
	RegisterCodec(new(JsonCodec))
}

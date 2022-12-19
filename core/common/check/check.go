package check

import (
	"github.com/nyan233/littlerpc/core/middle/codec"
	lreflect "github.com/nyan233/littlerpc/internal/reflect"
	"reflect"
)

// MarshalFromUnsafe structPtr == nil时不会返回空结果, 而是构造一个*interface{}传给
// Codec做为反序列化的类型, 这是反序列化出的具体类型由特定的Codec进行选择, 有些要求一定要具体类型的serialization
// 可能会Panic, 比如protobuf
// data中没数据时返回nil结果, 而不关心structPtr中的数据
func MarshalFromUnsafe(codec codec.Codec, data []byte, value interface{}) (interface{}, error) {
	if data == nil || len(data) == 0 {
		return nil, nil
	}
	var marshalValue interface{}
	if value == nil {
		marshalValue = new(interface{})
		value = interface{}(1)
	} else {
		marshalValue, _ = lreflect.ToTypePtr(value)
	}
	err := codec.Unmarshal(data, marshalValue)
	if err != nil {
		return nil, err
	}
	// 指针类型和非指针类型的返回值不同
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		return value, nil
	} else {
		return reflect.ValueOf(marshalValue).Elem().Interface(), nil
	}
}

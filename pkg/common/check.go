package common

import (
	"errors"
	lreflect "github.com/nyan233/littlerpc/internal/reflect"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"reflect"
)

func CheckCoderType(codec codec.Codec, data []byte, structPtr interface{}) (interface{}, error) {
	if structPtr == nil || data == nil || len(data) == 0 {
		return nil, errors.New("no satisfy unmarshal case")
	}
	val, _ := lreflect.ToTypePtr(structPtr)
	err := codec.Unmarshal(data, val)
	if err != nil {
		return nil, err
	}
	// 指针类型和非指针类型的返回值不同
	if reflect.TypeOf(structPtr).Kind() == reflect.Ptr {
		return structPtr, nil
	} else {
		return reflect.ValueOf(val).Elem().Interface(), nil
	}
}

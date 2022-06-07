package reflect

import (
	"reflect"
	"unsafe"
)

// IdentifyTypeNoInfo 无类型信息的识别，识别一些简单类型的长度并返回它们的指针
// 只支持32-64位长度的类型
func IdentifyTypeNoInfo(value interface{}) (unsafe.Pointer, int) {
	inter := (*Eface)(unsafe.Pointer(&value))
	switch value.(type) {
	case int32, uint32, float32:
		return inter.data, 4
	case int64, uint64, float64:
		return inter.data, 8
	default:
		return nil, 0
	}
}

// IdentArrayOrSliceType 识别数组/切片元素类型
// type&data为nil则直接返回nil，有类型信息则新建再识别
func IdentArrayOrSliceType(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	typ := reflect.TypeOf(value)
	if !(typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array) {
		panic("value type is not a slice or array")
	}
	val := reflect.ValueOf(value)
	// 为0则动态创建再判断类型
	if val.Len() == 0 {
		val = reflect.MakeSlice(typ, 1, 1)
	}
	return val.Index(0).Interface()
}

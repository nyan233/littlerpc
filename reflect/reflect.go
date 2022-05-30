package reflect

import (
	"reflect"
	"unsafe"
)

type Eface struct {
	typ unsafe.Pointer
	data unsafe.Pointer
}

// IdentifyTypeNoInfo 无类型信息的识别，识别一些简单类型的长度并返回它们的指针
// 只支持32-64位长度的类型
func IdentifyTypeNoInfo(value interface{}) (unsafe.Pointer,int) {
	inter := (*Eface)(unsafe.Pointer(&value))
	switch value.(type) {
	case int32,uint32,float32:
		return inter.data,4
	case int64,uint64,float64:
		return inter.data,8
	default:
		return nil,0
	}
}

// IdentArrayOrSliceType 识别数组的类型
func IdentArrayOrSliceType(value interface{}) interface{} {
	typ := reflect.TypeOf(value)
	if !(typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array) {
		panic("value type is not a slice or array")
	}
	val := reflect.ValueOf(value)
	// 为0则动态创建再判断类型
	if val.Len() == 0 {
		val = reflect.New(typ)
	}
	return val.Index(0).Interface()
}
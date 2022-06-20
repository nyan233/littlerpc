package reflect

import (
	"reflect"
	"unsafe"
)

// 转换系列函数

// ToTypePtr 将一个非指针的interface{}转换为他的指针底层表示
// Bool表示是否非指针类型
func ToTypePtr(v interface{}) (interface{}, bool) {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		return v, false
	}
	// Map/Chan 的eface data指针是双重指针(**hmap)，要做特殊处理
	if typ.Kind() == reflect.Map {
		ptr := (*[2]unsafe.Pointer)(unsafe.Pointer(&v))[1]
		inter := typeToEfaceNoNew(reflect.PtrTo(typ), nil)
		return *(*interface{})(unsafe.Pointer(&Eface{
			typ:  (*[2]unsafe.Pointer)(unsafe.Pointer(&inter))[0],
			data: unsafe.Pointer(&ptr),
		})), true
	}
	return reflect.New(reflect.PtrTo(typ)).Interface(), true
}

// ToValueTypeEface 如果reflect.Value为nil则返回可以和nil比较的interface{}
func ToValueTypeEface(val reflect.Value) interface{} {
	eface := (*[2]uintptr)(unsafe.Pointer(&val))
	if eface[0] == 0 && eface[1] == 0 {
		return nil
	}
	return val.Interface()
}

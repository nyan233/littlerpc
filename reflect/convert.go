package reflect

import (
	"reflect"
	"unsafe"
)

// 转换系列函数

// ToTypePtr 将一个非指针的interface{}转换为他的指针底层表示
func ToTypePtr(v interface{}) interface{} {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		return v
	}
	// Map/Chan 的eface data指针是双重指针，要做特殊处理
	if val.Kind() == reflect.Map {
		ptr := (*[2]unsafe.Pointer)(unsafe.Pointer(&v))[1]
		inter := typeToEfaceNoNew(reflect.PtrTo(val.Type()),nil)
		return *(*interface{})(unsafe.Pointer(&Eface{
			typ:  (*[2]unsafe.Pointer)(unsafe.Pointer(&inter))[0],
			data: unsafe.Pointer(&ptr),
		}))
	}
	return typeToEfaceNoNew(reflect.PtrTo(val.Type()),v)
}

// ToValueTypeEface 如果reflect.Value为nil则返回可以和nil比较的interface{}
func ToValueTypeEface(val reflect.Value) interface{} {
	eface := (*[2]uintptr)(unsafe.Pointer(&val))
	if eface[0] == 0 && eface[1] == 0 {
		return nil
	}
	return val.Interface()
}
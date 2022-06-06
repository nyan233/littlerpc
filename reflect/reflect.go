package reflect

import (
	"reflect"
	"unsafe"
)

type Eface struct {
	typ unsafe.Pointer
	data unsafe.Pointer
}

type Iface struct {
	itab unsafe.Pointer
	data unsafe.Pointer
}


// FuncInputTypeList 返回函数的输入参数类型列表，空接口切片表示
func FuncInputTypeList(value reflect.Value) []interface{} {
	typ := value.Type()
	typs := make([]interface{},typ.NumIn())
	for k := range typs {
		if typ.Kind() == reflect.Interface {
			typs[k] = value.Interface()
		}
		typs[k] = typeToEfaceNew(typ.In(k))
	}
	return typs
}

// FuncOutputTypeList 返回函数的返回值类型列表，空接口切片表示
func FuncOutputTypeList(value reflect.Value) []interface{} {
	typ := value.Type()
	typs := make([]interface{},typ.NumOut())
	for k := range typs {
		if typ.Kind() == reflect.Interface {
			typs[k] = value.Interface()
		}
		typs[k] = typeToEfaceNew(typ.Out(k))
	}
	return typs
}

// 将reflect.Type中携带的类型信息转换为efce的类型信息
// 会重新创建数据并修正eface data指针
func typeToEfaceNew(typ reflect.Type) interface{} {
	iface := (*[2]unsafe.Pointer)(unsafe.Pointer(&typ))
	// Map/Chan 的eface data指针是双重指针，要做特殊处理
	if typ.Kind() == reflect.Map {
		return *(*interface{})(unsafe.Pointer(&Eface{
			typ:  iface[1],
			data: unsafe.Pointer(reflect.MakeMap(typ).Pointer()),
		}))
	}
	return *(*interface{})(unsafe.Pointer(&Eface{
		typ:  iface[1],
		data: unsafe.Pointer(reflect.New(typ).Pointer()),
	}))
}

// 将reflect.Type中携带的类型信息转换为efce的类型信息
// 不会会重新创建数据
func typeToEfaceNoNew(typ reflect.Type,val interface{}) interface{} {
	iface := (*[2]unsafe.Pointer)(unsafe.Pointer(&typ))
	// Map/Chan 的eface data指针是双重指针，要做特殊处理
	if typ.Kind() == reflect.Map {
		return *(*interface{})(unsafe.Pointer(&Eface{
			typ:  iface[1],
			data: unsafe.Pointer(reflect.ValueOf(val).Pointer()),
		}))
	}
	return *(*interface{})(unsafe.Pointer(&Eface{
		typ:  iface[1],
		data: (*[2]unsafe.Pointer)(unsafe.Pointer(&val))[1],
	}))
}

// InterDataPointer 获得val对应eface-data指针的值
func InterDataPointer(val interface{}) unsafe.Pointer {
	return (*Eface)(unsafe.Pointer(&val)).data
}
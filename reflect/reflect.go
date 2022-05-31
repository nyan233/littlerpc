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

// FuncInputTypeList 返回函数的输入参数类型列表，空接口切片表示
func FuncInputTypeList(value reflect.Value) []interface{} {
	typ := value.Type()
	typs := make([]interface{},typ.NumIn())
	for k := range typs {
		typs[k] = typeToEfaceNew(typ.In(k))
	}
	return typs
}

// FuncOutputTypeList 返回函数的返回值类型列表，空接口切片表示
func FuncOutputTypeList(value reflect.Value) []interface{} {
	typ := value.Type()
	typs := make([]interface{},typ.NumOut())
	for k := range typs {
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

func InterDataPointer(val interface{}) unsafe.Pointer {
	return (*Eface)(unsafe.Pointer(&val)).data
}
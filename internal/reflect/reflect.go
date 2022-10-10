package reflect

import (
	"reflect"
	"unsafe"
)

type Eface struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

type Iface struct {
	itab unsafe.Pointer
	data unsafe.Pointer
}

func RealType(value reflect.Value) reflect.Value {
	if value.Kind() != reflect.Interface {
		return value
	}
	return RealType(reflect.ValueOf(value.Interface()))
}

// FuncInputTypeList 返回函数的输入参数类型列表，空接口切片表示
// directNew是一个过滤器回调函数, 返回true指示new的类型是其原始类型, 返回false则代表new一个非指针类型并返回其指针
// directNew()其入参的index当(isRecv == true), 它的值相当于减去接收器的参数的长度
//
//	Example
//	if directNew() == true { Input(*reflect.Value) Exec -> return new(*reflect.Value)}
//	if directNew() == false { Input(*reflect.Value) Exec -> tmp := new(reflect.Value) -> return &tmp}
func FuncInputTypeList(value reflect.Value, start int, isRecv bool, directNew func(i int) bool) []interface{} {
	typ := value.Type()
	typs := make([]interface{}, 0, typ.NumIn())
	if isRecv {
		start++
	}
	if isRecv && cap(typs) <= 1 {
		return nil
	}
	inputIndex := -1
	for i := start; i < cap(typs); i++ {
		inputIndex++
		if directNew != nil && directNew(inputIndex) {
			typs = append(typs, reflect.New(typ.In(i)).Interface())
			continue
		}
		if typ.In(i).Kind() == reflect.Interface {
			typs = append(typs, reflect.New(typ.In(i)).Interface())
			continue
		}
		// 非指针的类型
		if typ.In(i).Kind() != reflect.Ptr {
			typs = append(typs, reflect.New(typ.In(i)).Elem().Interface())
			continue
		}
		typs = append(typs, reflect.New(typ.In(i).Elem()).Interface())
	}
	return typs
}

// FuncOutputTypeList 返回函数的返回值类型列表，空接口切片表示
// directNew是一个过滤器回调函数, 返回true指示new的类型是其原始类型, 返回false则代表new一个非指针类型并返回其指针
// directNew()其入参的index当(isRecv == true), 它的值相当于减去接收器的参数的长度
//
//	Example
//	if directNew() == true { Input(*reflect.Value) Exec -> return new(*reflect.Value)}
//	if directNew() == false { Input(*reflect.Value) Exec -> tmp := new(reflect.Value) -> return &tmp}
func FuncOutputTypeList(value reflect.Value, directNew func(i int) bool) []interface{} {
	typ := value.Type()
	typs := make([]interface{}, 0, typ.NumOut())
	if cap(typs) == 0 {
		return nil
	}
	for i := 0; i < cap(typs); i++ {
		if directNew != nil && directNew(i) {
			typs = append(typs, reflect.New(typ.Out(i)).Elem().Interface())
			continue
		}
		if typ.Out(i).Kind() == reflect.Interface {
			typs = append(typs, reflect.New(typ.Out(i)).Interface())
			continue
		}
		// 非指针类型
		if typ.Out(i).Kind() != reflect.Ptr {
			typs = append(typs, reflect.New(typ.Out(i)).Elem().Interface())
			continue
		}
		typs = append(typs, reflect.New(typ.Out(i).Elem()).Interface())
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
func typeToEfaceNoNew(typ reflect.Type, val interface{}) interface{} {
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

// DeepEqualNotType 主要用于比较
func DeepEqualNotType(x, y interface{}) bool {
	if x == nil && y == nil {
		return true
	} else if x == nil || y == nil {
		return false
	}
	xValue := reflect.ValueOf(x)
	yValue := reflect.ValueOf(y)
	if xValue.Type() == reflect.TypeOf([]interface{}{0}) || yValue.Type() == reflect.TypeOf([]interface{}{0}) {
		if xValue.Kind() != reflect.Slice || yValue.Kind() != reflect.Slice {
			return false
		}
		return deepEqualNotTypeOnArray(xValue, yValue)
	} else if xValue.Type() != yValue.Type() {
		return false
	} else {
		return reflect.DeepEqual(x, y)
	}
}

func deepEqualNotTypeOnArray(x, y reflect.Value) bool {
	if x.Len() != y.Len() {
		return false
	}
	var rangeN reflect.Value
	var cmpN reflect.Value
	if _, xOK := x.Interface().([]interface{}); xOK {
		rangeN = x
		cmpN = y
	} else {
		rangeN = y
		cmpN = x
	}
	length := x.Len()
	for i := 0; i < length; i++ {
		rangeV := RealType(rangeN.Index(i))
		cmpV := RealType(cmpN.Index(i))
		if rangeV.Kind() == reflect.Slice && cmpV.Kind() == reflect.Slice {
			deepEqualNotTypeOnArray(rangeV, cmpV)
			continue
		}
		if !reflect.DeepEqual(rangeV.Interface(), cmpV.Interface()) {
			return false
		}
	}
	return true
}

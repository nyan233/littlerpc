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

// FuncInputTypeListReturnValue 返回函数的输入参数类型列表，空接口切片表示
// directNew是一个过滤器回调函数, 返回true指示new的类型是其原始类型, 返回false则根据是否指针类型来返回不同的数据
//
//	Example
//	if directNew() == true { Input(*reflect.Value) Exec -> return new(*reflect.Value)}
//	if directNew() == false { Input(reflect.Value) Exec -> tmp := new(reflect.Value) -> return *tmp}
//	if directNew() == false { Input(*reflect.Value) Exec -> tmp := new(reflect.Value) -> return tmp}
func FuncInputTypeListReturnValue(tList []reflect.Type, start int, directNew func(i int) bool, skipInter bool) []reflect.Value {
	if (tList != nil && len(tList) == 0) || start >= len(tList) {
		return nil
	}
	result := make([]reflect.Value, 0, len(tList)-start)
	for index, typ := range tList {
		if index < start {
			continue
		}
		if directNew != nil && directNew(index) {
			result = append(result, reflect.New(typ).Elem())
			continue
		}
		if typ.Kind() == reflect.Interface {
			if skipInter {
				result = append(result, reflect.ValueOf(nil))
			} else {
				result = append(result, reflect.New(typ))
			}
			continue
		}
		// 非指针的类型
		if typ.Kind() != reflect.Ptr {
			result = append(result, reflect.New(typ).Elem())
			continue
		}
		result = append(result, reflect.New(typ.Elem()))
	}
	return result
}

// FuncOutputTypeList 返回函数的返回值类型列表，空接口切片表示
// directNew是一个过滤器回调函数, 返回true指示new的类型是其原始类型, 返回false则根据是否指针类型来返回不同的数据
//
//	Example
//	if directNew() == true { Input(*reflect.Value) Exec -> return new(*reflect.Value)}
//	if directNew() == false { Input(reflect.Value) Exec -> tmp := new(reflect.Value) -> return *tmp}
//	if directNew() == false { Input(*reflect.Value) Exec -> tmp := new(reflect.Value) -> return tmp}
func FuncOutputTypeList(tList []reflect.Type, directNew func(i int) bool, skipInter bool) []interface{} {
	if tList != nil && len(tList) == 0 {
		return nil
	}
	result := make([]interface{}, 0, len(tList))
	for index, typ := range tList {
		if directNew != nil && directNew(index) {
			result = append(result, reflect.New(typ).Elem().Interface())
			continue
		}
		if typ.Kind() == reflect.Interface {
			if skipInter {
				result = append(result, nil)
			} else {
				result = append(result, reflect.New(typ).Interface())
			}
			continue
		}
		// 非指针类型
		if typ.Kind() != reflect.Ptr {
			result = append(result, reflect.New(typ).Elem().Interface())
			continue
		}
		result = append(result, reflect.New(typ.Elem()).Interface())
	}
	return result
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

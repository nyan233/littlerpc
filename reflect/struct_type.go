package reflect

import (
	"reflect"
	"sync"
	"unsafe"
)

var (
	// Any.Any的非Slice类型的缓存
	cacheType sync.Map
	// Any.Any的Slice类型缓存
	cacheSliceType sync.Map
)

// CreateAnyStructOnElemType 通过数组/切片的元素类型
// 创建Any字段类型为[]val.Type的struct
// 返回的是其结构体的指针类型
func CreateAnyStructOnElemType(val interface{}) interface{} {
	eface := (*[2]uintptr)(unsafe.Pointer(&val))
	var typ reflect.Type
	// 查询有无缓存，缓存中有对应的*type结构则不需要创建
	typVal,ok := cacheSliceType.Load(eface[0])
	if !ok {
		typ = reflect.StructOf([]reflect.StructField{
			{Name: "Any", Type: reflect.SliceOf(reflect.TypeOf(val))},
		})
		cacheSliceType.Store(eface[0],typ)
	} else {
		typ = typVal.(reflect.Type)
	}
	ptrTyp := reflect.PtrTo(typ)
	sVal := reflect.New(typ).Interface()
	return typeToEfaceNoNew(ptrTyp, sVal)
}

// CreateAnyStructOnType 通过val的类型
// 创建Any字段类型为val.Type的struct,返回值类型为结构体的指针
func CreateAnyStructOnType(val interface{}) interface{} {
	eface := (*[2]uintptr)(unsafe.Pointer(&val))
	var typ reflect.Type
	// 查询有无缓存，缓存中有对应的*type结构则不需要创建
	typVal,ok := cacheType.Load(eface[0])
	if !ok {
		typ = reflect.StructOf([]reflect.StructField{
			{Name: "Any", Type: reflect.TypeOf(val)},
		})
		cacheType.Store(eface[0],typ)
	} else {
		typ = typVal.(reflect.Type)
	}
	ptrTyp := reflect.PtrTo(typ)
	sVal := reflect.New(typ).Interface()
	return typeToEfaceNoNew(ptrTyp, sVal)
}

// 装配Any.Any的空接口表示
func ComposeStructAnyEface(val interface{}, rawType reflect.Type) interface{} {
	eface := (*Eface)(unsafe.Pointer(&val))
	return typeToEfaceNoNew(rawType, *(*interface{})(unsafe.Pointer(&Eface{
		typ:  (*[2]unsafe.Pointer)(unsafe.Pointer(&rawType))[1],
		data: eface.data,
	})))
}

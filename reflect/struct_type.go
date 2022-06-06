package reflect

import "reflect"

// CreateAnyStructOnElemType 通过数组/切片的元素类型
// 创建Any字段类型为[]val.Type的struct
// 返回的是其结构体的指针类型
func CreateAnyStructOnElemType(val interface{}) interface{} {
	typ := reflect.StructOf([]reflect.StructField{
		{Name: "Any",Type: reflect.SliceOf(reflect.TypeOf(val))},
	})
	ptrTyp := reflect.PtrTo(typ)
	sVal := reflect.New(typ).Interface()
	return typeToEfaceNoNew(ptrTyp,sVal)
}

// CreateAnyStructOnType 通过val的类型
// 创建Any字段类型为val.Type的struct,返回值类型为结构体的指针
func CreateAnyStructOnType(val interface{}) interface{} {
	typ := reflect.StructOf([]reflect.StructField{
		{Name: "Any",Type: reflect.TypeOf(val)},
	})
	ptrTyp := reflect.PtrTo(typ)
	sVal := reflect.New(typ).Interface()
	return typeToEfaceNoNew(ptrTyp,sVal)
}

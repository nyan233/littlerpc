package reflect

import "reflect"

// CreateAnyStruct val是带有数组类型信息的eface
// 返回的是其结构体的指针类型
func CreateAnyStruct(val interface{}) interface{} {
	typ := reflect.StructOf([]reflect.StructField{
		{Name: "Any",Type: reflect.SliceOf(reflect.TypeOf(val))},
	})
	ptrTyp := reflect.PtrTo(typ)
	sVal := reflect.New(typ).Interface()
	return typeToEfaceNoNew(ptrTyp,sVal)
}
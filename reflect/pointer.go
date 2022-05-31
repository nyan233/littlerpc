package reflect

import "reflect"

// 关于一些指针的工具函数

// PtrSetValue 给指针设置对应的值
func PtrSetValue(ptrI interface{},val interface{}) interface{} {
	return typeToEfaceNoNew(reflect.TypeOf(ptrI),val)
}

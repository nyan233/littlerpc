package reflect

import "reflect"

// 关于一些指针的工具函数

// PtrDeriveValue 根据ptrI提供的type,val提供的data派生一个eface
// 不会修改ptrI中的数据，因为只使用了reflect.TypeOf()获取type的指针
// 并没有使用ptrI中对应的efce的data的指针
func PtrDeriveValue(ptrI interface{}, val interface{}) interface{} {
	// NoNew不会重新分配eface-data的结构内存
	return typeToEfaceNoNew(reflect.TypeOf(ptrI), val)
}

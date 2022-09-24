package common

import "reflect"

type ElemMeta struct {
	// instance type
	Typ reflect.Type
	// instance pointer
	Data reflect.Value
	// instance method collection
	Methods map[string]reflect.Value
}

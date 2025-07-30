package server

import "reflect"

func allocValueFromArgsType(args []reflect.Type, filter func(idx int) bool, newFn func(p reflect.Type) reflect.Value) []reflect.Value {
	res := make([]reflect.Value, len(args))
	for idx, arg := range args {
		if !filter(idx) {
			continue
		}
		if newFn != nil {
			res[idx] = newFn(arg)
		}
	}
	return res
}

func allocDefaultNewFunc(arg reflect.Type) reflect.Value {
	// 默认new行为
	// 非指针的类型
	if arg.Kind() != reflect.Ptr {
		return reflect.New(arg).Elem()
	} else {
		return reflect.New(arg.Elem())
	}
}

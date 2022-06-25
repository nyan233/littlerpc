package reflect

import (
	"reflect"
	"testing"
	"unsafe"
)

type testStruct struct {
	u1 uintptr
}

func TestFuncReflect(t *testing.T) {
	testFunc := func(i1 []int, i2 *testStruct) (*int, *testStruct) {
		return nil, nil
	}
	// 测试函数的参数列表
	typs := FuncInputTypeList(reflect.ValueOf(testFunc), false)
	_ = typs[0].([]int)
	_ = typs[1].(*testStruct)
	// 测试函数的返回值列表
	typs = FuncOutputTypeList(reflect.ValueOf(testFunc), false)
	_ = typs[0].(*int)
	_ = typs[1].(*testStruct)
}

func TestTypeTo(t *testing.T) {
	eface := typeToEfaceNew(reflect.TypeOf(new(int64)))
	_ = eface.(*int64)
	// Map Pointer
	eface = typeToEfaceNew(reflect.TypeOf(new(map[string]int)))
	_ = eface.(*map[string]int)
	// Map No Pointer
	eface = typeToEfaceNew(reflect.TypeOf(map[string]int{}))
	_ = eface.(map[string]int)

	// Type No New
	eface = typeToEfaceNoNew(reflect.TypeOf(*new(int)), 10)
	if eface != 10 {
		panic(interface{}("typeToEfaceNoNew return value failed"))
	}
}

func TestConcurrentTypeTo(t *testing.T) {

}

func TestReflectConv(t *testing.T) {
	i := typeToEfaceNew(reflect.TypeOf(new(int64)))
	_ = i.(*int64)
	v, _ := ToTypePtr(map[string]int{"hello": 1111})
	v = createMapPtr(map[int]int{1: 1})
	mapV := v.(*map[int]int)
	(*mapV)[1] = 2
}

func createMapPtr(val interface{}) interface{} {
	ptrTyp := reflect.PtrTo(reflect.TypeOf(val))
	eface := &Eface{}
	eface.typ = (*[2]unsafe.Pointer)(unsafe.Pointer(&ptrTyp))[1]
	var ptr = reflect.ValueOf(val).Pointer()
	eface.data = unsafe.Pointer(&ptr)
	return *(*interface{})(unsafe.Pointer(eface))
}

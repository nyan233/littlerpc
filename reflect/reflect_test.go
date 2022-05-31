package reflect

import (
	"reflect"
	"testing"
	"unsafe"
)

type testStruct struct {
	u1 uintptr
}

func testFunc(i1 []int,i2 *testStruct) {}

func TestFuncReflect(t *testing.T) {
	typs := FuncInputTypeList(reflect.ValueOf(testFunc))
	_ = typs[0].([]int)
	_ = typs[1].(*testStruct)
}

func TestReflectConv(t *testing.T) {
	i := typeToEfaceNew(reflect.TypeOf(new(int64)))
	t.Log(i)
	j := i.(*int64)
	t.Log(j)
	v := ToTypePtr(map[string]int{"hello":1111})
	t.Log(v)
	v = createMapPtr(map[int]int{1:1})
	mapV := v.(*map[int]int)
	(*mapV)[1] = 2
	t.Log(v)
}

func createMapPtr(val interface{}) interface{} {
	ptrTyp := reflect.PtrTo(reflect.TypeOf(val))
	eface := &Eface{}
	eface.typ = (*[2]unsafe.Pointer)(unsafe.Pointer(&ptrTyp))[1]
	var ptr = reflect.ValueOf(val).Pointer()
	eface.data = unsafe.Pointer(&ptr)
	return *(*interface{})(unsafe.Pointer(eface))
}
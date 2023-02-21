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
	// 测试函数的参数列表
	typs := FuncInputTypeListReturnValue([]reflect.Type{
		reflect.TypeOf([]int(nil)),
		reflect.TypeOf((*testStruct)(nil)),
	}, 0, nil, false)
	_ = typs[0].Interface().([]int)
	_ = typs[1].Interface().(*testStruct)
	// 测试函数的返回值列表
	typs2 := FuncOutputTypeList([]reflect.Type{
		reflect.TypeOf((*int)(nil)),
		reflect.TypeOf((*testStruct)(nil)),
	}, func(i int) bool {
		return true
	}, true)
	_ = typs2[0].(*int)
	_ = typs2[1].(*testStruct)
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

func TestDeepEqualNotType(t *testing.T) {
	normalCmp := []interface{}{
		123, 123,
		"str1", "str1",
		"str2", "str2",
		map[string]int{"heheda": 123},
		map[string]int{"heheda": 123},
	}
	for i := 0; i < len(normalCmp); i += 2 {
		if !DeepEqualNotType(normalCmp[i], normalCmp[i+1]) {
			t.Fatal("DeepEqualNotType normalCmp is not equal, index == ", i)
		}
	}
	sliceCmp1 := []interface{}{
		[]interface{}{1, 2, 3, 4},
		[]int{1, 2, 3, 4},
		[]interface{}{"s1", "s2", "s3"},
		[]string{"s1", "s2", "s3"},
		[]interface{}{"s1", 123, []interface{}{"hehe", "haha"}},
		[]interface{}{"s1", 123, []string{"hehe", "haha"}},
	}
	for i := 0; i < len(sliceCmp1); i += 2 {
		if !DeepEqualNotType(sliceCmp1[i], sliceCmp1[i+1]) {
			t.Fatal("DeepEqualNotType sliceCmp1 is not equal, index == ", i)
		}
	}
	sliceCmp2 := []interface{}{
		[]interface{}{1, "heheda", "lalala", "wahaha"},
		[]interface{}{1, "heheda", "lalala", "wahaha"},
		[]interface{}{map[string]int{"map1": 123}, 123, "hehe"},
		[]interface{}{map[string]int{"map1": 123}, 123, "hehe"},
		[]interface{}{123, "sss", "ssr", []interface{}{123, "234", 456, "789"}},
		[]interface{}{123, "sss", "ssr", []interface{}{123, "234", 456, "789"}},
		[]interface{}{123, "sss", "ssr3", []interface{}{123, "234", 456, "789"}}, // 判错
		[]interface{}{123, "sss", "ssr4", []interface{}{123, "234", 456, "789"}}, // 判错
	}
	for i := 0; i < len(sliceCmp2)-2; i += 2 {
		if !DeepEqualNotType(sliceCmp2[i], sliceCmp2[i+1]) {
			t.Fatal("DeepEqualNotType sliceCmp2 is not equal, index == ", i)
		}
	}
	if DeepEqualNotType(SliceIndex(sliceCmp2, -1), SliceIndex(sliceCmp2, -2)) {
		t.Fatal("DeepEqualNotType sliceCmp2 is not equal")
	}
}

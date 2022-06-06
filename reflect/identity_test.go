package reflect

import "testing"

// 测试类型识别系列函数
func TestTypeIdentify(t *testing.T) {
	_,tLen := IdentifyTypeNoInfo(uint64(0))
	if tLen != 8 {
		t.Fatal("identify type failed")
	}
	arrayT := IdentArrayOrSliceType(*new([]byte))
	_ = arrayT.(byte)
	_ = IdentArrayOrSliceType(nil)
}


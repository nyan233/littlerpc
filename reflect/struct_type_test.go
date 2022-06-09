package reflect

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestCreateSliceType(t *testing.T) {
	getTypeInfo(func(k int, v interface{}) {
		t.Run(fmt.Sprintf("CreateNoSliceType-%s",reflect.TypeOf(v).Kind()), func(t *testing.T) {
			for i := 0; i < 10; i++ {
				val := CreateAnyStructOnType(v)
				ComposeStructAnyEface(val,reflect.TypeOf(v))
			}
		})
	})
	getTypeInfo(func(k int, v interface{}) {
		t.Run(fmt.Sprintf("CreateSliceType-%s",reflect.TypeOf(v).Kind()), func(t *testing.T) {
			for i := 0; i < 10; i++ {
				val := CreateAnyStructOnElemType(v)
				ComposeStructAnyEface(val,reflect.TypeOf(v))
			}
		})
	})
}

func BenchmarkCrateAndJsonUnmarshal(b *testing.B) {
	type Any struct {
		Any interface{}
	}
	const JsonData = `{"Any": 1024}`
	b.Run("No-Reflect-Create", func(b *testing.B) {
		b.ReportAllocs()
		typ := reflect.TypeOf(*new(int64))
		for i := 0; i < b.N; i++ {
			val := CreateAnyStructOnType(*new(int64))
			_ = json.Unmarshal([]byte(JsonData), val)
			_ = ComposeStructAnyEface(val, typ)
		}
	})
	b.Run("Reflect-Create", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var any Any
			any.Any = *new(int64)
			_ = json.Unmarshal([]byte(JsonData), &any)
			_ = any.Any
		}
	})
}

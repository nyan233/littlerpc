package reflect

import (
	"reflect"
	"testing"
)

func TestConvert(t *testing.T) {
	nTyp, _ := ToTypePtr(*new(int))
	_ = nTyp.(*int)
	efceT := interface{}(10)
	val := ToValueTypeEface(reflect.ValueOf(efceT))
	if val != efceT {
		t.Fatal("efceT no equal value")
	}
}
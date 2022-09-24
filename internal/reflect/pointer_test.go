package reflect

import "testing"

func TestPointer(t *testing.T) {
	ptr := new(int)
	*ptr = 1 << 10
	ptr2 := PtrDeriveValue(ptr, 10).(*int)
	if *ptr2 != 10 {
		panic(interface{}("*ptr2 value is no correct"))
	}
}

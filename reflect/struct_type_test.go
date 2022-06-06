package reflect

import "testing"

func TestCreateSliceType(t *testing.T) {
	val := CreateAnyStructOnElemType(new(string))
	_ = val.(*struct{
		Any []*string
	})
	val = CreateAnyStructOnType(*new([]string))
	_ = val.(*struct{
		Any []string
	})
}

package reflect

// 该函数用于生成各种测试需要的类型信息
func getTypeInfo(fn func(k int,v interface{})) {
	fn(0,*new(string))
	fn(1,*new(int32))
	fn(2,*new(int64))
	fn(3,*new(uint32))
	fn(4,*new(uint64))
	fn(5,*new(float32))
	fn(6,*new(float64))
	fn(7,make(map[string]interface{}))
	fn(8,struct {
		Id int
		Name string
	}{})
}

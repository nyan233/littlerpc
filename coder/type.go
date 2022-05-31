package coder

type Type uint8

const (
	String   Type = (iota + 1) << 1
	Byte          // 1B
	Long          // 4B
	Integer       // 8B
	ULong         // 4B
	UInteger      // 8B
	Float         // 4B
	Double        // 8B
	Array
	Struct
	Map
	Pointer // 表示一个指针或者暗含一个指针
	Interface
)

// AnyArgs 用于所有的参数传递
// Map等类型需要标注类型参数
type AnyArgs struct {
	Any interface{}
}

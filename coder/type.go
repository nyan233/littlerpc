package coder

type Type uint8

const (
	String      Type = (iota + 1) << 1
	Boolean          // 1B
	Byte             // 1B
	Long             // 4B
	Integer          // 8B
	ULong            // 4B
	UInteger         // 8B
	Float            // 4B
	Double           // 8B
	Array            // 在go里表示数组和切片
	Struct           // 用于表示class/struct
	Map              // 通用的Map类型
	Pointer          // 表示一个指针或者暗含一个指针
	Interface        // 通用的接口类型
	ServerError      // 用于附加类型，用于区别是正常返回的错误还是Server遇到比如解析Json失败传回的错误
)

// AnyArgs 用于所有的参数传递
// Map等类型需要标注类型参数
type AnyArgs struct {
	Any interface{}
}

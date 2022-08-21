package error

// 定义LittleRpc内部会使用到的错误码

type Code int

func (c Code) String() string {
	return mappingStr[c]
}

const (
	Success               = 200  // 成功返回
	Unknown               = 300  // 用户过程返回了错误,但不是LittleRpc可以识别的错误
	MethodNoRegister      = 1404 // 需要调用的方法未被注册
	InstanceNoRegister    = 2404 // 需要调用的实例未被注册
	MessageDecodingFailed = 1020 // 载荷消息解码失败
	ServerError           = 500  // 服务器的其它错误
	CallArgsTypeErr       = 1040 // 过程的调用参数类型错误
	CodecMarshalErr       = 1050 // Codec在序列化数据时出错
	UnsafeOption          = 2060
)

var mappingStr = map[Code]string{
	Success:               "Success",
	Unknown:               "Unknown",
	MethodNoRegister:      "MethodNoRegister",
	InstanceNoRegister:    "InstanceNoRegister",
	MessageDecodingFailed: "MessageDecodingFailed",
	ServerError:           "ServerError",
	CallArgsTypeErr:       "CallArgsTypeErr",
	CodecMarshalErr:       "CodecMarshalErr",
	UnsafeOption:          "UnsafeOption",
}

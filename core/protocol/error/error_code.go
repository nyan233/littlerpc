package error

import (
	"fmt"
	"github.com/nyan233/littlerpc/core/utils/convert"
)

// 定义LittleRpc内部会使用到的错误码

type Code int

func (c Code) String() string {
	bytes, _ := c.MarshalJSON()
	return convert.BytesToString(bytes)
}

func (c Code) MarshalJSON() ([]byte, error) {
	codeStr, ok := mappingStr[c]
	// 用户自定义的错误玛
	if !ok {
		return convert.StringToBytes(fmt.Sprintf("\"Custom(%d)\"", c)), nil
	}
	return convert.StringToBytes(codeStr), nil
}

const (
	Success               = 200  // 成功返回
	Unknown               = 730  // 用户过程返回了错误,但不是LittleRpc可以识别的错误
	ServiceNotFound       = 750  // 需要调用的服务不存在
	MessageDecodingFailed = 780  // 载荷消息解码失败
	MessageEncodingFailed = 1060 // 载荷消息编码失败
	ServerError           = 690  // 服务器的其它错误
	ClientError           = 580  // 客户端产生的错误
	// CallArgsTypeErr TODO: 计划删除, v0.2.0时代的遗留产物
	CallArgsTypeErr = 1030 // 过程的调用参数类型错误
	CodecMarshalErr = 1050 // Codec在序列化数据时出错
	ConnectionErr   = 1070 // 连接错误
	ContextNotFound = 1080 // 要取消的context不存在
	UnsafeOption    = 2060 // 不安全的选项, 通常在服务器需要的东西没有准备好时触发
)

// NOTE: 不要尝试修改这个表，这个表不应该在运行时被改变或者被使用到
// NOTE: Little-Rpc的用户代码改变
var mappingStr = map[Code]string{
	Success:               "\"Success\"",
	Unknown:               "\"Unknown\"",
	ServiceNotFound:       "\"ServiceNotFound\"",
	MessageDecodingFailed: "\"MessageDecodingFailed\"",
	MessageEncodingFailed: "\"MessageEncodingFailed\"",
	ServerError:           "\"ServerError\"",
	ClientError:           "\"ClientError\"",
	CallArgsTypeErr:       "\"CallArgsTypeErr\"",
	CodecMarshalErr:       "\"CodecMarshalErr\"",
	ConnectionErr:         "\"ConnectionErr\"",
	ContextNotFound:       "\"ContextNotFound\"",
	UnsafeOption:          "\"UnsafeOption\"",
}

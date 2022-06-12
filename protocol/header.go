package protocol

const (
	MessageCall   string = "call"
	MessageReturn string = "return"
	MessagePing   string = "ping"
	MessagePong   string = "pong"

	DefaultEncodingType = "text"
	DefaultCodecType = "json"
)

// Header 对于有效载荷消息的头描述
// int/uint数值统一使用大端序
type Header struct {
	// call/return & ping/pong
	MsgType string
	// default text/gzip
	Encoding string
	// default json
	CodecType string
	// 要调用的方法名
	MethodName string
	// 消息ID，用于跟踪等用途
	MsgId uint64
	// 生成该消息的时间戳,精确到毫秒
	Timestamp uint64
}

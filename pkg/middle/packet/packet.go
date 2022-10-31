package packet

// Encoder 负责解包/拆包的实例要实现的接口
// 获得数据是经过json encode之后的Body,Header不会被发送到此处理
// 默认实现text&gzip
type Encoder interface {
	Scheme() string
	// EnPacket 装包
	EnPacket(p []byte) ([]byte, error)
	// UnPacket 拆包
	UnPacket(p []byte) ([]byte, error)
}

var (
	encoderCollection = make(map[string]Encoder, 8)
)

// TextPacket 注意:text类型的压缩器只是提供给client&server
// 的一个提示,client&server的代码应该针对此特殊处理，真实调用会导致panic
type TextPacket struct{}

func (t TextPacket) Scheme() string {
	return "text"
}

func (t TextPacket) EnPacket(p []byte) ([]byte, error) {
	panic(interface{}("text packet not able call"))
}

func (t TextPacket) UnPacket(p []byte) ([]byte, error) {
	panic(interface{}("text packet not able call"))
}

// RegisterEncoder 该调用是线程安全的
func RegisterEncoder(p Encoder) {
	if p == nil {
		panic("encoder is nil")
	}
	if p.Scheme() == "" {
		panic("encoder scheme is empty")
	}
	encoderCollection[p.Scheme()] = p
}

// GetEncoder 该调用是线程安全的
func GetEncoder(scheme string) Encoder {
	return encoderCollection[scheme]
}

func init() {
	RegisterEncoder(new(TextPacket))
	RegisterEncoder(new(GzipPacket))
}

package packet

import "sync"

// Encoder 负责解包/拆包的实例要实现的接口
// 获得数据是经过json encode之后的Body,Header不会被发送到此处理
// 默认实现text&gzip
type Encoder interface {
	Scheme() string
	// EnPacket 装包
	EnPacket(p []byte) ([]byte,error)
	// UnPacket 拆包
	UnPacket(p []byte) ([]byte,error)
}

var (
	packetCollection sync.Map
)

type TextPacket struct {}

func (t TextPacket) Scheme() string {
	return "text"
}

func (t TextPacket) EnPacket(p []byte) ([]byte,error) {
	return p,nil
}

func (t TextPacket) UnPacket(p []byte) ([]byte,error) {
	return p,nil
}

func RegisterEncoder(p Encoder) {
	packetCollection.Store(p.Scheme(),p)
}

func GetEncoder(scheme string) Encoder {
	encoder,ok := packetCollection.Load(scheme)
	if !ok {
		return nil
	}
	return encoder.(Encoder)
}

func init() {
	RegisterEncoder(new(TextPacket))
	RegisterEncoder(new(GzipPacket))
}

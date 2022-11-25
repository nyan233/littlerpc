package packer

// Packer 负责解包/拆包的实例要实现的接口
// 获得数据是经过json encode之后的Body,Header不会被发送到此处理
// 默认实现text&gzip
type Packer interface {
	Scheme() string
	// EnPacket 装包
	EnPacket(p []byte) ([]byte, error)
	// UnPacket 拆包
	UnPacket(p []byte) ([]byte, error)
}

var (
	packerCollection = make(map[string]Packer, 8)
)

// Text 注意:text类型的压缩器只是提供给client&server
// 的一个提示,client&server的代码应该针对此特殊处理，真实调用会导致panic
type Text struct{}

func (t Text) Scheme() string {
	return "text"
}

func (t Text) EnPacket(p []byte) ([]byte, error) {
	panic(interface{}("text packet not able call"))
}

func (t Text) UnPacket(p []byte) ([]byte, error) {
	panic(interface{}("text packet not able call"))
}

// Register 该调用是线程安全的
func Register(p Packer) {
	if p == nil {
		panic("encoder is empty")
	}
	if p.Scheme() == "" {
		panic("encoder scheme is empty")
	}
	packerCollection[p.Scheme()] = p
}

// Get 该调用是线程安全的
func Get(scheme string) Packer {
	return packerCollection[scheme]
}

func init() {
	Register(new(Text))
	Register(new(Gzip))
}

package packet

import "sync"

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
	manager = &encoderManager{
		packetSchemeCollection: map[string]Wrapper{},
		packetIndexCollection:  []Wrapper{},
	}
)

type encoderManager struct {
	mu                     sync.Mutex
	packetSchemeCollection map[string]Wrapper
	packetIndexCollection  []Wrapper
}

func (e *encoderManager) registerEncoder(encoder Encoder) {
	e.mu.Lock()
	defer e.mu.Unlock()
	wrapper := newEncoderWrapper(len(e.packetIndexCollection), encoder)
	e.packetSchemeCollection[wrapper.Scheme()] = wrapper
	e.packetIndexCollection = append(e.packetIndexCollection, wrapper)
}

func (e *encoderManager) getCodecFromScheme(scheme string) Wrapper {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.packetSchemeCollection[scheme]
}

func (e *encoderManager) getCodecFromIndex(index int) Wrapper {
	e.mu.Lock()
	defer e.mu.Unlock()
	// 这使得这个过程不会因为index超出了长度而panic
	if index >= len(e.packetIndexCollection) {
		return nil
	}
	return e.packetIndexCollection[index]
}

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
	manager.registerEncoder(p)
}

// GetEncoderFromScheme 该调用是线程安全的
func GetEncoderFromScheme(scheme string) Wrapper {
	return manager.getCodecFromScheme(scheme)
}

// GetEncoderFromIndex 该调用是线程安全的,且可以安全的使用任何索引数值
func GetEncoderFromIndex(index int) Wrapper {
	return manager.getCodecFromIndex(index)
}

func init() {
	RegisterEncoder(new(TextPacket))
	RegisterEncoder(new(GzipPacket))
}

package codec

import (
	"encoding/json"
	"sync"
)

type Codec interface {
	Scheme() string
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

var (
	manager = &codecManager{
		codecCollection:      map[string]Wrapper{},
		indexCodecCollection: []Wrapper{},
	}
)

type codecManager struct {
	mu                   sync.Mutex
	codecCollection      map[string]Wrapper
	indexCodecCollection []Wrapper
}

func (m *codecManager) registerCodec(c Codec) {
	m.mu.Lock()
	defer m.mu.Unlock()
	wrapper := newCodecWrapper(len(m.indexCodecCollection), c)
	m.codecCollection[c.Scheme()] = wrapper
	m.indexCodecCollection = append(m.indexCodecCollection, wrapper)
}

func (m *codecManager) getCodecFromScheme(scheme string) Wrapper {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.codecCollection[scheme]
}

func (m *codecManager) getCodecFromIndex(index int) Wrapper {
	m.mu.Lock()
	defer m.mu.Unlock()
	if index >= len(m.indexCodecCollection) {
		return nil
	}
	return m.indexCodecCollection[index]
}

// RegisterCodec 该调用是线程安全的
func RegisterCodec(c Codec) {
	manager.registerCodec(c)
}

// GetCodecFromScheme 该调用是线程安全的
func GetCodecFromScheme(scheme string) Wrapper {
	return manager.getCodecFromScheme(scheme)
}

// GetCodecFromIndex 该调用是线程安全的,且可以安全的使用任何数值来作为索引
func GetCodecFromIndex(index int) Wrapper {
	return manager.getCodecFromIndex(index)
}

type JsonCodec struct{}

func (j JsonCodec) Scheme() string {
	return "json"
}

func (j JsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j JsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func init() {
	RegisterCodec(new(JsonCodec))
}

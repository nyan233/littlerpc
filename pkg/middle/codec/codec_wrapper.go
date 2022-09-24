package codec

// Wrapper Codec的包装器
// 主要提供给client&server标注其Codec在一个抽象的表中的位置
type Wrapper interface {
	Scheme() string
	Index() int
	Instance() Codec
}

type codecWrapper struct {
	index int
	codec Codec
}

func (c *codecWrapper) Scheme() string {
	return c.codec.Scheme()
}

func (c *codecWrapper) Index() int {
	return c.index
}

func (c *codecWrapper) Instance() Codec {
	return c.codec
}

func newCodecWrapper(index int, codec Codec) Wrapper {
	return &codecWrapper{
		index: index,
		codec: codec,
	}
}

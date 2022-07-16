package packet

// Wrapper 同middle.Codec功能一致
type Wrapper interface {
	Scheme() string
	Index() int
	Instance() Encoder
}

type encoderWrapper struct {
	index   int
	encoder Encoder
}

func (e *encoderWrapper) Scheme() string {
	return e.encoder.Scheme()
}

func (e *encoderWrapper) Index() int {
	return e.index
}

func (e *encoderWrapper) Instance() Encoder {
	return e.encoder
}

func newEncoderWrapper(index int, encoder Encoder) *encoderWrapper {
	return &encoderWrapper{
		index:   index,
		encoder: encoder,
	}
}

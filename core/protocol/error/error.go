package error

type LErrorDesc interface {
	Code() int
	Message() string
	AppendMore(more interface{})
	Mores() []interface{}
	MarshalMores() ([]byte, error)
	UnmarshalMores([]byte) error
	error
}

type LErrors interface {
	// LNewErrorDesc 用于生产LittleRpc中的标准错误
	LNewErrorDesc(code int, message string, mores ...interface{}) LErrorDesc
	// LWarpErrorDesc 用于包装LittleRpc中的标准错误
	LWarpErrorDesc(desc LErrorDesc, mores ...interface{}) LErrorDesc
}

type LNewErrorDesc func(code int, message string, mores ...interface{}) LErrorDesc

type LWarpErrorDesc func(desc LErrorDesc, mores ...interface{}) LErrorDesc

package error

type LErrorDesc interface {
	GetCode() int
	SetCode(code int)
	GetMessage() string
	SetMessage(message string)
	AppendMore(more interface{})
	SetMores(mores []interface{})
	GetMores() []interface{}
	MarshalMores() ([]byte, error)
	UnmarshalMores([]byte) error
	error
}

type LBaseError struct {
	Code    int
	Message string
	Mores   []interface{}
}

func LNewBaseError(code int, message string, mores ...interface{}) *LBaseError {
	return &LBaseError{
		Code:    code,
		Message: message,
		Mores:   mores,
	}
}

type LNewErrorDesc func(code int, message string, mores ...interface{}) LErrorDesc

package protocol

type Error struct {
	Info  string
	Trace string
	More  []interface{}
}

func (e Error) Error() string {
	return e.Info + " " + e.Trace
}

func NewError(info string, trace string, more ...interface{}) *Error {
	return &Error{
		Info:  info,
		Trace: trace,
		More:  more,
	}
}

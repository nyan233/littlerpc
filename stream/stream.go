package stream

type Stream interface {
	ReadMessage(data interface{})
	WriteMessage(data interface{})
	Close(port int) error
}

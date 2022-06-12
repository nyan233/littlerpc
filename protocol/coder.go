package protocol

type Body struct {
	// 栈帧->远程
	Frame []FrameMd
}

type Message struct {
	Header Header
	Body   Body
}

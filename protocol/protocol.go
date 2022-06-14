package protocol

type Body struct {
	// 栈帧->远程
	Frame []FrameMd
}

type Message struct {
	Header Header
	Body   Body
}

// FrameMd 远程栈帧元数据
type FrameMd struct {
	ArgType Type
	// 附加类型
	AppendType Type
	Data        []byte
}

func (f *FrameMd) Encode(codec Codec,i interface{}) error {
	marshal, err := codec.Marshal(i)
	if err != nil {
		return err
	}
	f.Data = marshal
	return nil
}

func (f *FrameMd) Decode(codec Codec,i interface{}) error {
	return codec.Unmarshal(f.Data,i)
}

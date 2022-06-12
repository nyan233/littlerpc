package protocol

import "encoding/json"

// FrameMd 远程栈帧元数据
type FrameMd struct {
	ArgType Type
	// 附加类型
	AppendType Type
	Data        []byte
}

func (c *FrameMd) Encode(i interface{}) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.Data = data
	return nil
}

func (c *FrameMd) Decode(i interface{}) error {
	return json.Unmarshal(c.Data,i)
}

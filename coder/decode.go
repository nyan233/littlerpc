package coder

import "encoding/json"

// CalleeMd 被调用者(服务端)的元数据
type CalleeMd struct {
	ArgType Type
	MethodName string
	Rep []byte
}

func (c *CalleeMd) EncodeResponse(i interface{}) error {
	rep,err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.Rep = rep
	return nil
}

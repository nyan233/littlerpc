package coder

import "encoding/json"

// CallerMd 客户端(Caller)请求调用的元数据
type CallerMd struct {
	// Meta Type
	ArgType Type
	// 附加类型
	AppendType Type
	// Request Args
	Req []byte
}

func (c *CallerMd) EncodeRequest(i interface{}) error {
	req,err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.Req= req
	return nil
}

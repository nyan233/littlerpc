package common

import (
	"github.com/nyan233/littlerpc/protocol"
)

func BufferIoEncodeMessage(msg *protocol.Message) []byte{
	var data []byte
	data = append(data,msg.EncodeHeader()...)
	for _,v := range msg.Body {
		data = append(data,v...)
	}
	return data
}
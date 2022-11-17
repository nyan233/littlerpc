package msgparser

import (
	"encoding/json"
	"errors"

	"github.com/nyan233/littlerpc/pkg/common/jsonrpc2"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/protocol/message"
)

type jsonRpc2Handler struct {
	Codec codec.Codec
}

func (j *jsonRpc2Handler) Header() []byte {
	return []byte{jsonrpc2.Header}
}

func (j *jsonRpc2Handler) BaseLen() (BaseLenType, int) {
	return SingleRequest, -1
}

func (j *jsonRpc2Handler) MessageLength(base []byte) int {
	panic("implement me")
}

func (j *jsonRpc2Handler) Unmarshal(data []byte, msg *message.Message) (Action, error) {
	var request jsonrpc2.Request
	err := j.Codec.Unmarshal(data, &request)
	if err != nil {
		return -1, err
	}
	if request.Version != jsonrpc2.Version {
		return -1, errors.New("hash")
	}
	if request.Codec == "" {
		msg.MetaData.Store(message.CodecScheme, message.DefaultCodec)
	} else {
		msg.MetaData.Store(message.CodecScheme, j.Codec.Scheme())
	}
	// jsonrpc2不支持压缩编码
	msg.MetaData.Store(message.PackerScheme, message.DefaultPacker)
	msg.SetMsgId(uint64(request.Id))
	if request.MetaData != nil {
		for k, v := range request.MetaData {
			msg.MetaData.Store(k, v)
		}
	}
	if request.Method == "" {
		return -1, errors.New("hash")
	}
	msg.SetServiceName(request.Method)
	if request.Params == nil || len(request.Params) == 0 {
		return UnmarshalComplete, nil
	}
	switch request.Params[0] {
	case '[':
		var msgs []json.RawMessage
		err := j.Codec.Unmarshal(request.Params, &msgs)
		if err != nil {
			return -1, err
		}
		for _, v := range msgs {
			msg.AppendPayloads(v)
		}
	default:
		msg.AppendPayloads(request.Params)
	}
	msg.GetAndSetLength()
	return UnmarshalComplete, nil
}

package msgparser

import (
	"encoding/json"
	"errors"
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/nyan233/littlerpc/protocol/message"
	"strings"
)

const (
	JsonRPC2KeepAlive = "rpc.keepalive"
	JsonRPC2Version   = "2.0"
)

type JsonRPC2CallDesc struct {
	Version  string            `json:"jsonrpc"`
	Method   string            `json:"method"`
	Codec    string            `json:"rpc_codec"`
	MetaData map[string]string `json:"rpc_metadata"`
	Id       int64             `json:"id"`
	Params   []byte            `json:"params"`
}

type JsonRpc2Handler struct {
	Codec codec.Codec
}

func (j *JsonRpc2Handler) BaseLen() (BaseLenType, int) {
	return SingleRequest, -1
}

func (j *JsonRpc2Handler) MessageLength(base []byte) int {
	//TODO implement me
	panic("implement me")
}

func (j *JsonRpc2Handler) Unmarshal(data []byte, msg *message.Message) (Action, error) {
	var callDesc JsonRPC2CallDesc
	err := j.Codec.Unmarshal(data, &callDesc)
	if err != nil {
		return -1, err
	}
	if callDesc.Version != JsonRPC2Version {
		return -1, errors.New("hash")
	}
	if callDesc.Codec == "" {
		msg.SetCodecType(message.DefaultCodecType)
	} else {
		msg.SetCodecType(uint8(codec.GetCodecFromScheme(callDesc.Codec).Index()))
	}
	// jsonrpc2不支持压缩编码
	msg.SetEncoderType(message.DefaultEncodingType)
	msg.SetMsgId(uint64(callDesc.Id))
	if callDesc.MetaData != nil {
		for k, v := range callDesc.MetaData {
			msg.MetaData.Store(k, v)
		}
	}
	if callDesc.Method == "" {
		return -1, errors.New("hash")
	}
	methodSplit := strings.Split(callDesc.Method, ".")
	if len(methodSplit) != 2 {
		return -1, errors.New("hash")
	}
	msg.SetInstanceName(methodSplit[0])
	msg.SetMethodName(methodSplit[1])
	if callDesc.Params == nil || len(callDesc.Params) == 0 {
		return UnmarshalComplete, nil
	}
	switch callDesc.Params[0] {
	case '[':
		var msgs []json.RawMessage
		err := j.Codec.Unmarshal(callDesc.Params, &msgs)
		if err != nil {
			return -1, err
		}
		for _, v := range msgs {
			msg.AppendPayloads(v)
		}
	default:
		msg.AppendPayloads(callDesc.Params)
	}
	msg.GetAndSetLength()
	return UnmarshalComplete, nil
}

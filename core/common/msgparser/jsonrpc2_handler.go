package msgparser

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/nyan233/littlerpc/core/common/jsonrpc2"
	"github.com/nyan233/littlerpc/core/middle/codec"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils/convert"
	"math"
	"strconv"
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
	var base jsonrpc2.BaseMessage
	err := j.Codec.Unmarshal(data, &base)
	if err != nil {
		return -1, err
	}
	if base.Version != jsonrpc2.Version {
		return -1, errors.New("unknown message version")
	}
	if base.MessageType > math.MaxUint8 {
		return -1, errors.New("message type overflow")
	}
	msg.SetMsgId(base.Id)
	msg.SetMsgType(uint8(base.MessageType))
	if base.MetaData != nil {
		for k, v := range base.MetaData {
			msg.MetaData.Store(k, v)
		}
	}
	isJsonParams := true
	var packerScheme string
	if base.MetaData != nil {
		packerScheme = base.MetaData[message.PackerScheme]
		codecScheme := base.MetaData[message.CodecScheme]
		isJsonParams = codecScheme == "" || codecScheme == message.DefaultCodec
	}
	if !(packerScheme == "" || packerScheme == message.DefaultPacker) {
		return -1, errors.New("jsonrpc2 not supported only text packer")
	}
	switch uint8(base.MessageType) {
	case message.ContextCancel, message.Call:
		var trait jsonrpc2.RequestTrait
		err = j.Codec.Unmarshal(data, &trait)
		if err != nil {
			return -1, err
		}
		msg.SetServiceName(trait.Method)
		if trait.Params == nil || len(trait.Params) == 0 {
			return UnmarshalComplete, nil
		}
		switch trait.Params[0] {
		case '[':
			var msgs []json.RawMessage
			err = j.Codec.Unmarshal(trait.Params, &msgs)
			if err != nil {
				return -1, err
			}
			for _, v := range msgs {
				var bytes []byte
				var err error
				if !isJsonParams {
					bytes, err = base64.StdEncoding.DecodeString(convert.BytesToString(v))
					if err != nil {
						return 0, err
					}
				} else {
					bytes = v
				}
				msg.AppendPayloads(bytes)
			}
		default:
			msg.AppendPayloads(trait.Params)
		}
		msg.GetAndSetLength()
	case message.Return:
		var trait jsonrpc2.ResponseTrait
		err = j.Codec.Unmarshal(data, &trait)
		if err != nil {
			return -1, err
		}
		if trait.Error != nil {
			msg.MetaData.Store(message.ErrorCode, strconv.Itoa(trait.Error.Code))
			msg.MetaData.Store(message.ErrorMessage, trait.Error.Message)
			if trait.Error.Data != nil || len(trait.Error.Data) != 0 {
				msg.MetaData.Store(message.ErrorMore, convert.BytesToString(trait.Error.Data))
			}
		}
		for _, result := range trait.Result {
			msg.AppendPayloads(result)
		}
	default:
		return -1, errors.New("unknown message type")
	}
	return UnmarshalComplete, nil
}

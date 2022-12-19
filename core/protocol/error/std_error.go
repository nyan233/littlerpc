package error

import (
	"encoding/json"
	"github.com/nyan233/littlerpc/core/utils/convert"
)

type LStdError struct {
	LCode    Code          `json:"code"`
	LMessage string        `json:"message"`
	LMores   []interface{} `json:"mores"`
}

func LNewStdError(code int, message string, mores ...interface{}) LErrorDesc {
	return &LStdError{
		LCode:    Code(code),
		LMessage: message,
		LMores:   mores,
	}
}

func LWarpStdError(desc LErrorDesc, mores ...interface{}) LErrorDesc {
	return &LStdError{
		LCode:    Code(desc.Code()),
		LMessage: desc.Message(),
		LMores:   append(desc.Mores(), mores...),
	}
}

func (L *LStdError) Code() int {
	return int(L.LCode)
}

func (L *LStdError) Message() string {
	return L.LMessage
}

func (L *LStdError) AppendMore(more interface{}) {
	L.LMores = append(L.LMores, more)
}

func (L *LStdError) Mores() []interface{} {
	return L.LMores
}

func (L *LStdError) Error() string {
	bytes, err := json.Marshal(L)
	if err != nil {
		panic("json.Marshal failed : " + err.Error())
	}
	return convert.BytesToString(bytes)
}

func (L *LStdError) MarshalMores() ([]byte, error) {
	return json.Marshal(L.LMores)
}

func (L *LStdError) UnmarshalMores(bytes []byte) error {
	return json.Unmarshal(bytes, &L.LMores)
}

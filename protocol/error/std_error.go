package error

import (
	"encoding/json"
	"fmt"
	"strings"
)

type LStdError struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Mores   []interface{} `json:"mores"`
}

func LNewStdError(code int, message string, mores ...interface{}) LErrorDesc {
	return &LStdError{
		Code:    code,
		Message: message,
		Mores:   mores,
	}
}

func (L *LStdError) GetCode() int {
	return L.Code
}

func (L *LStdError) SetCode(code int) {
	L.Code = code
}

func (L *LStdError) GetMessage() string {
	return L.Message
}

func (L *LStdError) SetMessage(message string) {
	L.Message = message
}

func (L *LStdError) AppendMore(more interface{}) {
	L.Mores = append(L.Mores, more)
}

func (L *LStdError) SetMores(mores []interface{}) {
	L.Mores = mores
}

func (L *LStdError) GetMores() []interface{} {
	return L.Mores
}

func (L *LStdError) Error() string {
	var sb strings.Builder
	sb.Grow(64)
	fmt.Fprintf(&sb, `{"Code":%s, `, Code(L.Code).String())
	fmt.Fprintf(&sb, `"Message":%s, `, L.Message)
	if L.Mores == nil || len(L.Mores) == 0 {
		return sb.String()
	}
	sb.WriteString(`"Mores":`)
	bytes, err := json.Marshal(L.Mores)
	if err != nil {
		panic("json.Marshal failed : " + err.Error())
	}
	sb.Write(bytes)
	sb.WriteString("}")
	return sb.String()
}

func (L *LStdError) MarshalMores() ([]byte, error) {
	return json.Marshal(L.Mores)
}

func (L *LStdError) UnmarshalMores(bytes []byte) error {
	return json.Unmarshal(bytes, L)
}

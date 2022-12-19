package errorhandler

import (
	"encoding/json"
	"fmt"
	error2 "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrorHandler(t *testing.T) {
	eh := NewStackTrace()
	const (
		NCall    = 32
		OneSendN = 4
	)
	var err error2.LErrorDesc
	for i := 0; i < NCall; i++ {
		sendErrInfo := make([]interface{}, 0, OneSendN)
		for j := 0; j < OneSendN; j++ {
			sendErrInfo = append(sendErrInfo, fmt.Sprintf("test%d", j))
		}
		if err == nil {
			err = eh.LNewErrorDesc(error2.ConnectionErr, "Connection", sendErrInfo...)
		} else {
			err = eh.LWarpErrorDesc(err, sendErrInfo...)
		}
	}
	assert.NotNil(t, err)
	assert.Equal(t, len(err.Mores()), NCall*OneSendN)
	type ErrorType struct {
		Mores []interface{} `json:"mores"`
	}
	var errValue ErrorType
	assert.Equal(t, json.Unmarshal([]byte(err.Error()), &errValue), nil)
	assert.NotNil(t, errValue.Mores)
	for k, v := range errValue.Mores {
		stack, ok := v.(map[string]interface{})
		if k == len(errValue.Mores)-1 && !ok {
			t.Fatal("error information no stack trace")
		}
		if !ok {
			continue
		}
		stackTrace, ok := stack["stack"]
		if !ok {
			t.Fatal("lookup stack map ok buf no stack trace")
		}
		for _, stackTraceValue := range stackTrace.([]interface{}) {
			_, ok := stackTraceValue.(string)
			if !ok {
				t.Fatal("stack trace type not string")
			}
		}
	}
	oldLen := len(err.Mores())
	moresBytes, iErr := err.MarshalMores()
	assert.Nil(t, iErr)
	assert.Nil(t, err.UnmarshalMores(moresBytes))
	assert.Equal(t, len(err.Mores()), oldLen+1)
}

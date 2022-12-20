package main

import (
	"go/format"
	"testing"
)

func TestMethod(t *testing.T) {
	m := Method{
		Receive: Argument{
			Name: "p",
			Type: "TestIProxy",
		},
		ServiceName: "TestService",
		Name:        "Say",
		InputList: []Argument{
			{
				Name: "ctx",
				Type: "context.Context",
			},
			{
				Name: "userName",
				Type: "string",
			},
			{
				Name: "password",
				Type: "string",
			},
			{
				Name: "id",
				Type: "int64",
			},
		},
		OutputList: []Argument{
			{
				Name: "users",
				Type: "map[string]int64",
			},
			{
				Name: "err",
				Type: "error",
			},
		},
	}
	bytes, err := format.Source([]byte(m.FormatToSync()))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
}

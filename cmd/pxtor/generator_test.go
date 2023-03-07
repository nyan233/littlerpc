package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFutures(t *testing.T) {
	*dir = "./test"
	*receiver = "test.Test"
	*sourceName = "littlerpc/internal/test1"
	*style = SyncStyle
	*generateId = true
	for i := 0; i < 100; i++ {
		genCode()
	}
	*sourceName = "Test"
	genCode()
}

func TestGenApi(t *testing.T) {
	after, err := genSync(Argument{
		Name: "p",
		Type: "TestProxy",
	}, "Hello", "littlerpc/test/pxtor/internal", []Argument{
		{"s1", "string"}, {"d1", "int"},
	}, []Argument{
		{"", "string"}, {"", "error"},
	})
	assert.NotEqualf(t, after(), "", "result equal empty")
	assert.Nil(t, err)
	_, err = genAsyncApi("Hello", "littlerpc/test/pxtor/internal", "Add",
		[]string{"s1", "d1"}, []string{"string", "int"}, []string{"string", "error"})
	assert.Nil(t, err)
}

func TestCreateBeforeCode(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal(err)
		}
	}()
	createBeforeCode("main", "Hello1", "littlerpc/internal/test1", []string{
		"func(h *Hello1)Mehtod1(ctx context.Context) error",
		"func(h *Hello1)Mehtod2(ctx context.Context) error",
		"func(h *Hello1)Mehtod3(ctx context.Context) error",
	}, nil)
}

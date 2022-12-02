package main

import (
	"testing"
)

func TestFutures(t *testing.T) {
	*dir = "./test"
	*receiver = "test.Test"
	*sourceName = "littlerpc/internal/test1"
	for i := 0; i < 100; i++ {
		genCode()
	}
}

func TestGenApi(t *testing.T) {
	_, err := genSyncApi("Hello", "littlerpc/test/pxtor/internal", "Add",
		[]string{"s1", "d1"}, []string{"string", "int"}, []string{"string", "error"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = genAsyncApi("Hello", "littlerpc/test/pxtor/internal", "Add",
		[]string{"s1", "d1"}, []string{"string", "int"}, []string{"string", "error"})
	if err != nil {
		t.Fatal(err)
	}
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
	})
}

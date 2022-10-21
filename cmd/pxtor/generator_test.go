package main

import (
	"testing"
)

func TestFutures(t *testing.T) {
	*dir = "./test"
	*receiver = "test.Test"
	for i := 0; i < 100; i++ {
		genCode()
	}
}

func TestGenApi(t *testing.T) {
	_, err := genSyncApi("Hello", "Add", []string{"s1", "d1"}, []string{"string", "int"}, []string{"string", "error"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = genAsyncApi("Hello", "Add", []string{"s1", "d1"}, []string{"string", "int"}, []string{"string", "error"})
	if err != nil {
		t.Fatal(err)
	}
}

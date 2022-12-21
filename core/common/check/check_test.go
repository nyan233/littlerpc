package check

import (
	"github.com/nyan233/littlerpc/core/middle/codec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckCoderType(t *testing.T) {
	_, err := UnMarshalFromUnsafe(&codec.Json{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = UnMarshalFromUnsafe(&codec.Json{}, nil, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	bytes := []byte("{\"hello\":\"123\",\"dd\":\"456\"}")
	var testData map[string]string
	comparaData := map[string]string{
		"hello": "123",
		"dd":    "456",
	}
	uTestData, err := UnMarshalFromUnsafe(&codec.Json{}, bytes, testData)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, uTestData, comparaData)
	uTestData, err = UnMarshalFromUnsafe(&codec.Json{}, bytes, &testData)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, uTestData, &comparaData)
	uTestData, err = UnMarshalFromUnsafe(&codec.Json{}, bytes, nil)
	if err != nil {
		t.Fatal(err)
	}
}

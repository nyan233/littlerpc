package check

import (
	"github.com/nyan233/littlerpc/pkg/middle/codec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckCoderType(t *testing.T) {
	_, err := CheckCoderType(&codec.JsonCodec{}, nil, nil)
	if err == nil {
		t.Fatal("error equal nil")
	}
	bytes := []byte("{\"hello\":\"123\",\"dd\":\"456\"}")
	var testData map[string]string
	comparaData := map[string]string{
		"hello": "123",
		"dd":    "456",
	}
	uTestData, err := CheckCoderType(&codec.JsonCodec{}, bytes, testData)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, uTestData, comparaData)
	uTestData, err = CheckCoderType(&codec.JsonCodec{}, bytes, &testData)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, uTestData, &comparaData)
}

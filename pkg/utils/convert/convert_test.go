package convert

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvert(t *testing.T) {
	const Str = "1234567891234567891"
	assert.Equal(t, len(StringToBytes(Str)), len(Str))
	assert.Equal(t, cap(StringToBytes(Str)), len(Str))
	assert.Equal(t, len(BytesToString(StringToBytes(Str))), len(Str))
}

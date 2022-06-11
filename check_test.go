package littlerpc

import (
	"github.com/nyan233/littlerpc/coder"
	"testing"
)

func TestCheckCoderBaseType(t *testing.T) {
	for i := coder.String; i < coder.Type(100);i++{
		checkCoderBaseType(i)
	}
}

func TestCheckIType(t *testing.T) {
	for i := coder.String; i <= coder.ServerError;i+=2 {
		inter,_ := mappingCoderNoPtrType(i)
		_ = checkIType(inter)
		inter,_ = mappingCoderPtrType(i)
		_ = checkIType(inter)
	}
}
package littlerpc

import (
	"github.com/nyan233/littlerpc/protocol"
	"testing"
)

func TestCheckCoderBaseType(t *testing.T) {
	for i := protocol.String; i < protocol.Type(100);i++{
		checkCoderBaseType(i)
	}
}

func TestCheckIType(t *testing.T) {
	for i := protocol.String; i <= protocol.ServerError;i+=2 {
		inter,_ := mappingCoderNoPtrType(i)
		_ = checkIType(inter)
		inter,_ = mappingCoderPtrType(i)
		_ = checkIType(inter)
	}
}
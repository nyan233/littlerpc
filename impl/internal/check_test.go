package internal

import (
	"github.com/nyan233/littlerpc/protocol"
	"testing"
)

func TestCheckCoderBaseType(t *testing.T) {
	for i := protocol.String; i < protocol.Type(100);i++{
		CheckCoderBaseType(i)
	}
}

func TestCheckIType(t *testing.T) {
	for i := protocol.String; i <= protocol.ServerError;i+=2 {
		inter,_ := MappingCoderNoPtrType(i)
		_ = CheckIType(inter)
		inter,_ = MappingCoderPtrType(i)
		_ = CheckIType(inter)
	}
}
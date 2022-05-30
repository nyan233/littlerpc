package littlerpc

import "github.com/nyan233/littlerpc/coder"

//func fixJsonArrayType(i interface{},typ coder.Type) interface{} {
//
//}

func fixJsonType(i interface{},typ coder.Type) interface{} {
	eType, err := mappingReflectNoPtrType(typ,i)
	if err != nil {
		return nil
	}
	return eType
}

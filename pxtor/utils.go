package main

import (
	"errors"
	"fmt"
	"go/ast"
	"os"
	"strings"
)

func handleAstType(expr ast.Expr, file *os.File) string {
	switch expr.(type) {
	case *ast.MapType:
		typ := expr.(*ast.MapType)
		pos := fileSet.Position(typ.Pos())
		buf := []byte{0}
		_, err := file.ReadAt(buf, int64(pos.Offset))
		if err != nil {
			panic(interface{}(err))
		}
		if buf[0] == '*' {
			return "*map[" + handleAstType(typ.Key, file) + "]" + handleAstType(typ.Value, file)
		}
		return "map[" + handleAstType(typ.Key, file) + "]" + handleAstType(typ.Value, file)
	case *ast.Ident:
		ident := expr.(*ast.Ident)
		pos := fileSet.Position(ident.Pos())
		buf := []byte{0}
		_, err := file.ReadAt(buf, int64(pos.Offset))
		if err != nil {
			panic(interface{}(err))
		}
		if buf[0] == '*' {
			return "*" + ident.Name
		}
		return ident.Name
	case *ast.StarExpr:
		s := expr.(*ast.StarExpr)
		pos := fileSet.Position(s.Pos())
		buf := []byte{0}
		_, err := file.ReadAt(buf, int64(pos.Offset))
		if err != nil {
			panic(interface{}(err))
		}
		if buf[0] == '*' {
			return "*" + handleAstType(s.X, file)
		}
		return handleAstType(s.X, file)
	case *ast.ArrayType:
		at := expr.(*ast.ArrayType)
		pos := fileSet.Position(at.Pos())
		buf := []byte{0}
		_, err := file.ReadAt(buf, int64(pos.Offset))
		if err != nil {
			panic(interface{}(err))
		}
		if buf[0] == '*' {
			return "*[]" + handleAstType(at.Elt, file)
		}
		return "[]" + handleAstType(at.Elt, file)
	case *ast.SelectorExpr:
		se := expr.(*ast.SelectorExpr)
		return handleAstType(se.X, file) + "." + handleAstType(se.Sel, file)
	default:
		panic(interface{}("type is no supported"))
	}
}

func writeDefaultValue(str string) (string,error) {
	// 切片和指针类型直接返回nil
	if str[0] == '*' || (str[0] == '[' && str[1] == ']'){
		return "nil",nil
	}
	switch str {
	case "int8", "int16","uint16":
		return "",errors.New("no support type")
	case "uint8","byte","int32","int64","uint32","uint64","int":
		return "0",nil
	case "string":
		return `""`,nil
	case "bool":
		return "false",nil
	default:
		// littlerpc不支持chan类型的参数
		if str[:4] == "chan" {
			return "",errors.New("no support type")
		}
		// map类型也返回nil
		if str[:3] == "map" {
			return "nil",nil
		}
		// 是数组并非是切片
		if str[0] == '[' && str[1] != ']' {
			return str + "{}",nil
		}
		// 默认被设定为struct类型
		return str + "{}",nil
	}
}


// 根据方法集生成对应的接口
// Example : type dd interface {Say();Sel();}
func createInterface(typeName string,methods []string) string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "type %s interface {", typeName)
	for _,method := range methods {
		sb.WriteString(method)
		sb.WriteString(";")
	}
	return sb.String()
}

package main

import (
	"go/ast"
	"os"
)

func handleAstType(expr ast.Expr,file *os.File) string {
	switch expr.(type) {
	case *ast.MapType:
		typ := expr.(*ast.MapType)
		pos := fileSet.Position(typ.Pos())
		buf := []byte{0}
		_, err := file.ReadAt(buf, int64(pos.Offset))
		if err != nil {
			panic(err)
		}
		if buf[0] == '*' {
			return "*map[" + handleAstType(typ.Key,file) + "]" + handleAstType(typ.Value,file)
		}
		return "map[" + handleAstType(typ.Key,file) + "]" + handleAstType(typ.Value,file)
	case *ast.Ident:
		ident := expr.(*ast.Ident)
		pos := fileSet.Position(ident.Pos())
		buf := []byte{0}
		_, err := file.ReadAt(buf, int64(pos.Offset) - 1)
		if err != nil {
			panic(err)
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
			panic(err)
		}
		if buf[0] == '*' {
			return "*" + s.X.(*ast.Ident).Name
		}
		return s.X.(*ast.Ident).Name
	case *ast.ArrayType:
		at := expr.(*ast.ArrayType)
		pos := fileSet.Position(at.Pos())
		buf := []byte{0}
		_, err := file.ReadAt(buf, int64(pos.Offset))
		if err != nil {
			panic(err)
		}
		if buf[0] == '*' {
			return "*[]" + handleAstType(at.Elt,file)
		}
		return "[]" + handleAstType(at.Elt,file)
	default:
		panic("type is no supported")
	}
}

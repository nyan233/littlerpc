package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

var (
	receiver = flag.String("r", "", "代理对象的接收器: package.RecvName")
	dir      = flag.String("d", "./", "解析接收器的路径: ./")
	outName  = flag.String("o", "", "输出的文件名，默认的格式: receiver_proxy.go")
	fileSet  *token.FileSet
)

func main() {
	flag.Parse()
	if *receiver == "" {
		panic("no receiver specified")
	}

	tmp := strings.SplitN(*receiver, ".", 2)
	pkgName, recvName := tmp[0], tmp[1]
	// 输出文件名
	if *outName == "" {
		*outName = recvName + "_proxy.go"
	}
	// 打开文件
	_, err := os.OpenFile(*outName, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	fileSet = token.NewFileSet()
	parseDir, err := parser.ParseDir(fileSet, *dir, nil, 0)
	if err != nil {
		panic(err)
	}
	// ast.Print(fileSet,parseDir[pkgName].Files["test/proxy_2.go"])
	pkgDir := parseDir[pkgName]
	funcStrs := make([]string, 0, 20)
	for k, v := range pkgDir.Files {
		rawFile, err := os.Open("./" + k)
		if err != nil {
			panic(err)
		}
		tmp := getAllFunc(v, rawFile,func(recvT string) bool {
			if recvT == recvName {
				return true
			}
			return false
		})
		funcStrs = append(funcStrs, tmp...)
	}
}

func getAllFunc(file *ast.File, rawFile *os.File,filter func(recvT string) bool) []string {
	funcStrs := make([]string, 0)
	for _, v := range file.Decls {
		funcDecl, ok := v.(*ast.FuncDecl)
		if !ok {
			continue
		}
		receiver := funcDecl.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident)
		if !filter(receiver.Name) {
			continue
		}
		var sb strings.Builder
		sb.Grow(128)
		// funcStr
		sb.WriteString("func(proxy ")
		sb.WriteString(receiver.Name)
		sb.WriteString(") ")
		// 拼接函数名
		// func(proxy Test) Proxy2(
		sb.WriteString(funcDecl.Name.Name)
		sb.WriteString("(")
		// 参数名称的列表
		params := make([]string, 0, len(funcDecl.Type.Params.List))
		// 处理参数的序列化
		for _, pv := range funcDecl.Type.Params.List {
			// 多个参数同一类型的时候可能参数列表会是这样的: s1,s2 string
			// 这种情况要处理
			for _, pvName := range pv.Names {
				sb.WriteString(pvName.Name)
				params = append(params, pvName.Name)

				// 类型肯定只有一个，不可能多个参数多个类型
				sb.WriteString(" ")
				sb.WriteString(handleAstType(pv.Type,rawFile))
				sb.WriteString(",")
			}
		}
		// 处理参数列表的结束符
		sb.WriteString(") ")
		// result types
		rTypes := make([]string, 0, len(funcDecl.Type.Results.List))
		// 开始根据返回值类型注入littlerpc client的代码
		for _, rv := range funcDecl.Type.Results.List {
			rTypes = append(rTypes, handleAstType(rv.Type,rawFile))
		}
		// 返回值列表
		if len(rTypes) > 1 {
			sb.WriteByte('(')
		}
		for k, v := range rTypes {
			sb.WriteString(v)
			if k < len(rTypes)-1 {
				sb.WriteByte(',')
			}
		}
		if len(rTypes) > 1 {
			sb.WriteByte(')')
		}
		// inject call
		sb.WriteString(" {\n\t")
		sb.WriteString("inter := proxy.Call(")
		sb.WriteString(fmt.Sprintf("\"%s\",", funcDecl.Name.Name))
		for k, v := range params {
			sb.WriteString(v)
			if k < len(rTypes)-1 {
				sb.WriteByte(',')
			}
		}
		sb.WriteString(")\n\t")
		// inject function body
		for k, v := range rTypes {
			sb.WriteString(fmt.Sprintf("r%d := inter[%d].(%s)\n\t", k, k, v))
		}
		sb.WriteString("return ")
		for k := range rTypes {
			sb.WriteString("r")
			sb.WriteString(strconv.Itoa(k))
			if k < len(rTypes)-1 {
				sb.WriteByte(',')
			}
		}
		sb.WriteString("\n}\n")
		funcStrs = append(funcStrs, sb.String())
	}
	return funcStrs
}

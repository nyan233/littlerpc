package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
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
	genCode()
}

func genCode() {
	tmp := strings.SplitN(*receiver, ".", 2)
	pkgName, recvName := tmp[0], tmp[1]
	// 输出文件名
	if *outName == "" {
		*outName = recvName + "_proxy.go"
	}
	fileSet = token.NewFileSet()
	parseDir, err := parser.ParseDir(fileSet, *dir, nil, 0)
	if err != nil {
		panic(err)
	}
	// 等文件集解析完再打开文件
	file, err := os.OpenFile(*dir+"/"+*outName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	// ast.Print(fileSet,parseDir[pkgName].Files["test/proxy_2.go"])
	// 创建
	pkgDir := parseDir[pkgName]
	funcStrs := make([]string, 0, 20)
	// 要写入到文件的数据，提供这个是为了方便格式化生成的代码
	var fileBuffer bytes.Buffer
	fileBuffer.Grow(512)
	for k, v := range pkgDir.Files {
		rawFile, err := os.Open(path.Dir(*dir) + "/" + k)
		if err != nil {
			panic(err)
		}
		tmp := getAllFunc(v, rawFile, recvName+"Proxy", func(recvT string) bool {
			if recvT == recvName {
				return true
			}
			return false
		})
		funcStrs = append(funcStrs, tmp...)
	}
	fileBuffer.WriteString(createBeforeCode(pkgName, recvName+"Proxy",recvName+"Interface",funcStrs))
	for _, v := range funcStrs {
		fileBuffer.WriteString("\n")
		fileBuffer.WriteString(v)
	}
	if string(fileBuffer.Bytes()[fileBuffer.Len()-4:]) == "}\n}\n" {
		fmt.Println("double }")
	}
	fmtBytes, err := format.Source(fileBuffer.Bytes())
	if string(fmtBytes[len(fmtBytes)-4:]) == "}\n}\n" {
		fmt.Println("double }")
	}
	writeN, err := file.Write(fmtBytes)
	if err != nil {
		panic(err)
	}
	if writeN != len(fmtBytes) {
		panic(errors.New("write format bytes no equal"))
	}
}

func getAllFunc(file *ast.File, rawFile *os.File, proxyRecvName string, filter func(recvT string) bool) []string {
	funcStrs := make([]string, 0)
	for _, v := range file.Decls {
		funcDecl, ok := v.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if funcDecl.Recv == nil {
			continue
		}
		var receiver *ast.Ident
		for _, v := range funcDecl.Recv.List {
			// 目前只支持生成底层类型是struct的代理对象
			sExp, ok := v.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			ident, ok := sExp.X.(*ast.Ident)
			if !ok {
				continue
			}
			receiver = ident
		}
		// 无接收器的函数不是正确的声明
		if receiver == nil {
			continue
		}
		if !filter(receiver.Name) {
			continue
		}
		var sb strings.Builder
		sb.Grow(128)
		// funcStr
		sb.WriteString("func(proxy ")
		// 判断是否指针
		if handleAstType(receiver, rawFile)[0] == '*' {
			sb.WriteByte('*')
		}
		sb.WriteString(proxyRecvName)
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
				sb.WriteString(handleAstType(pv.Type, rawFile))
				sb.WriteString(",")
			}
		}
		// 处理参数列表的结束符
		sb.WriteString(") ")
		// result types
		rTypes := make([]string, 0, 4)
		if funcDecl.Type.Results == nil {
			goto handleBody
		}
		// 开始根据返回值类型注入littlerpc client的代码
		for _, rv := range funcDecl.Type.Results.List {
			rTypes = append(rTypes, handleAstType(rv.Type, rawFile))
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
	handleBody:
		// inject call
		sb.WriteString(" {\n\t")
		// littlerpc规定的合法的过程中至少需要一个error类型的返回值
		if funcDecl.Type.Results == nil {
			panic("generate function no return value")
		} else {
			sb.WriteString("inter,err := proxy.Call(")
		}
		sb.WriteString(fmt.Sprintf("\"%s.%s\",", receiver.Name, funcDecl.Name.Name))
		for _, v := range params {
			sb.WriteString(v)
			sb.WriteByte(',')
		}
		sb.WriteString(")\n\t")
		// 注入判断Call err的代码
		sb.WriteString("if err != nil { return ")
		for _,v := range rTypes[:len(rTypes) - 1] {
			s,err := writeDefaultValue(v)
			if err != nil {
				panic(err)
			}
			sb.WriteString(s)
			sb.WriteString(",")
		}
		sb.WriteString("err }\n\t")
		// inject function body
		for k, v := range rTypes {
			// error == nil时默认的断言会panic
			// 为了不panic必须使用带成功与否返回的类型断言
			if k == len(rTypes) - 1 {
				sb.WriteString(fmt.Sprintf("r%d,_ := inter[%d].(%s)\n\t",k,k,v))
				continue
			}
			sb.WriteString(fmt.Sprintf("r%d := inter[%d].(%s)\n\t", k, k, v))
		}
		sb.WriteString("return ")
		for k := range rTypes {
			// 返回值最后一个是error
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


// 在这里生成包注释、导入、工厂函数、各种需要的类型
func createBeforeCode(pkgName string, typeName,interName string,allFunc []string) string {
	var sb strings.Builder
	sb.Grow(1024)
	// 生成包注释
	fmt.Fprintf(&sb,"/*\n\t%-12s : littlerpc-generator", "@Generator")
	fmt.Fprintf(&sb,"\n\t%-12s : %s", "@CreateTime", time.Now().String())
	fmt.Fprintf(&sb,"\n\t%-12s : littlerpc-generator\n*/\n", "@Author")
	// 生成包名和导入文件
	fmt.Fprintf(&sb,"package " + pkgName + "\n")
	fmt.Fprintf(&sb,"import (\n\t")
	fmt.Fprintf(&sb,"\"github.com/nyan233/littlerpc/impl/client\"\n)\n")
	// 生成被代理对象方法集的接口描述
	fmt.Fprintf(&sb,"type %s interface {",interName)
	for _,v := range allFunc {
		// func (x receiver) Say(i int) error {...
		methodMeta := strings.SplitN(v,")",2)[1]
		methodMeta = strings.SplitN(methodMeta,"{",2)[0]
		sb.WriteString(methodMeta)
		sb.WriteString(";")
	}
	sb.WriteString("}\n\n")
	// 生成类型名和工厂函数
	fmt.Fprintf(&sb,"type %s struct {\n\t*client.Client\n}\n", typeName)
	fmt.Fprintf(&sb,"func New%s(client *client.Client) %s {", typeName, interName)
	fmt.Fprintf(&sb,"\n\tproxy := &%s{}\n\terr := client.BindFunc(proxy)\n\t", typeName)
	sb.WriteString("if err != nil {\n\tpanic(err)\n\t}\n\tproxy.Client = client\n\treturn proxy\n}\n")
	return sb.String()
}

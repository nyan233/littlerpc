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
		panic(interface{}("no receiver specified"))
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
		panic(interface{}(err))
	}
	// 等文件集解析完再打开文件
	file, err := os.OpenFile(*dir+"/"+*outName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		panic(interface{}(err))
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
			panic(interface{}(err))
		}
		tmp := getAllFunc(v, rawFile, func(recvT string) bool {
			if recvT == recvName {
				return true
			}
			return false
		})
		funcStrs = append(funcStrs, tmp...)
	}
	fileBuffer.WriteString(createBeforeCode(pkgName, recvName, recvName+"Proxy", recvName+"Interface", funcStrs))
	for _, v := range funcStrs {
		fileBuffer.WriteString("\n\n")
		fileBuffer.WriteString(v)
	}
	if string(fileBuffer.Bytes()[fileBuffer.Len()-4:]) == "}\n}\n" {
		fmt.Println("double }")
	}
	fmtBytes, err := format.Source(fileBuffer.Bytes())
	if err != nil {
		panic(err)
	}
	writeN, err := file.Write(fmtBytes)
	if err != nil {
		panic(interface{}(err))
	}
	if writeN != len(fmtBytes) {
		panic(interface{}(errors.New("write format bytes no equal")))
	}
}

func getAllFunc(file *ast.File, rawFile *os.File, filter func(recvT string) bool) []string {
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
		// 被代理对象的类型名
		recvName := receiver.Name
		// 被代理对应持有的方法名
		funName := funcDecl.Name.Name
		// 输入参数名字列表
		inNameList := make([]string, 0, 4)
		// 输入参数类型列表
		inTypeList := make([]string, 0, 4)
		// 输出参数类型列表
		outTypeList := make([]string, 0, 4)
		// 处理参数的序列化
		for _, pv := range funcDecl.Type.Params.List {
			// 多个参数同一类型的时候可能参数列表会是这样的: s1,s2 string
			// 这种情况要处理
			for _, pvName := range pv.Names {
				// 添加到输入参数名字列表
				inNameList = append(inNameList, pvName.Name)
				// 类型肯定只有一个，不可能多个参数多个类型
				// 添加到输入参数类型列表
				inTypeList = append(inTypeList, handleAstType(pv.Type, rawFile))
			}
		}
		// 找出所有的返回值类型
		for _, rv := range funcDecl.Type.Results.List {
			outTypeList = append(outTypeList, handleAstType(rv.Type, rawFile))
		}
		syncApi, err := genSyncApi(recvName, funName, inNameList, inTypeList, outTypeList)
		if err != nil {
			return nil
		}
		//asyncApi, err := genAsyncApi(recvName, funName, inNameList, inTypeList, outTypeList)
		//if err != nil {
		//	return nil
		//}
		funcStrs = append(funcStrs, syncApi)
		//funcStrs = append(funcStrs, asyncApi[0])
		//funcStrs = append(funcStrs, asyncApi[1])
	}
	return funcStrs
}

// 生成同步调用的Api
func genSyncApi(recvName, funName string, inNameList, inTypeList, outList []string) (string, error) {
	if len(inNameList) != len(inTypeList) {
		return "", errors.New("inNameList and inTypeList length not equal")
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "func (p %sProxy) %s(", recvName, funName)
	for i := 0; i < len(inNameList); i++ {
		fmt.Fprintf(&sb, "%s %s,", inNameList[i], inTypeList[i])
	}
	sb.WriteString(") ")
	for k, v := range outList {
		// 多返回值的情况
		if len(outList) > 1 && k == 0 {
			sb.WriteString("(")
		}
		sb.WriteString(v)
		// 不是最后一个返回值才添加分隔符
		if len(outList) > 1 && len(outList)-1 != k {
			sb.WriteString(",")
		}
		// 多返回值的情况
		if len(outList) > 1 && k == len(outList)-1 {
			sb.WriteString(")")
		}
	}
	if len(outList) > 1 {
		fmt.Fprintf(&sb, "{rep,err := p.Call(\"%s.%s\",", recvName, funName)
	} else {
		fmt.Fprintf(&sb, "{_,err := p.Call(\"%s.%s\",", recvName, funName)
	}
	for _, v := range inNameList {
		sb.WriteString(v)
		sb.WriteString(",")
	}
	sb.WriteString(");")
	for k, v := range outList[:len(outList)-1] {
		fmt.Fprintf(&sb, "r%d,_ := rep[%d].(%s);", k, k, v)
	}
	// 生成最终返回的代码
	sb.WriteString("return ")
	for k := range outList[:len(outList)-1] {
		if k == len(outList)-1 {
			fmt.Fprintf(&sb, "r%d", k)
			continue
		}
		fmt.Fprintf(&sb, "r%d,", k)
	}
	sb.WriteString("err;}")
	return sb.String(), nil
}

// 生成异步调用的Api
func genAsyncApi(recvName, funName string, inNameList, inTypeList, outList []string) (asyncApi [2]string, err error) {
	if len(inNameList) != len(inTypeList) {
		return [2]string{}, errors.New("inNameList and inTypeList length not equal")
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "func (p %sProxy) Async%s(", recvName, funName)
	for i := 0; i < len(inNameList); i++ {
		fmt.Fprintf(&sb, "%s %s,", inNameList[i], inTypeList[i])
	}
	fmt.Fprintf(&sb, ") error {return p.AsyncCall(\"%s.%s\",", recvName, funName)
	for _, v := range inNameList {
		sb.WriteString(v)
		sb.WriteByte(',')
	}
	sb.WriteString(")}")
	asyncApi[0] = sb.String()
	sb.Reset()
	fmt.Fprintf(&sb, "func (p %sProxy) Register%sCallBack(fn func(", recvName, funName)
	for k, v := range outList {
		fmt.Fprintf(&sb, "r%s %s,", strconv.Itoa(k), v)
	}
	sb.WriteString("))")
	fmt.Fprintf(&sb, "{p.RegisterCallBack(\"%s.%s\",func(rep []interface{}, err error) {", recvName, funName)
	// gen error check
	sb.WriteString("if err != nil {fn(")
	for k, v := range outList {
		// 关于error的生成必须独立处理，否则则会被替换为nil作为默认值
		if k == len(outList)-1 {
			// 一定要注入return,否则过程在出错的时候也会调用无错才会调用的回调函数
			sb.WriteString("err);return};")
			continue
		}
		str, err := writeDefaultValue(v)
		if err != nil {
			return [2]string{}, err
		}
		sb.WriteString(str)
		sb.WriteString(",")
	}
	// 生成断言的代码
	for k, v := range outList {
		// error类型的返回值使用安全断言
		if v == "error" {
			fmt.Fprintf(&sb, "r%d,_ := rep[%d].(%s);", k, k, v)
			continue
		}
		fmt.Fprintf(&sb, "r%d := rep[%d].(%s);", k, k, v)
	}
	// 最后生成调用的代码
	sb.WriteString("fn(")
	for k := range outList {
		fmt.Fprintf(&sb, "r%d,", k)
	}
	sb.WriteString(");})}")
	asyncApi[1] = sb.String()
	return
}

// 在这里生成包注释、导入、工厂函数、各种需要的类型
func createBeforeCode(pkgName, recvName string, typeName, interName string, allFunc []string) string {
	var sb strings.Builder
	sb.Grow(1024)
	// 生成包注释
	fmt.Fprintf(&sb, "/*\n\t%-12s : littlerpc-generator", "@Generator")
	fmt.Fprintf(&sb, "\n\t%-12s : %s", "@CreateTime", time.Now().String())
	fmt.Fprintf(&sb, "\n\t%-12s : littlerpc-generator", "@Author")
	fmt.Fprintf(&sb, "\n\t%-12s : code is auto generate do not edit\n*/\n", "@Comment")
	// 生成包名和导入文件
	fmt.Fprintf(&sb, "package "+pkgName+"\n")
	fmt.Fprintf(&sb, "import (\n\t")
	fmt.Fprintf(&sb, "\"github.com/nyan233/littlerpc/client\"\n)\n")
	// 生成被代理对象方法集的接口描述
	fmt.Fprintf(&sb, "type %s interface {", interName)
	for _, v := range allFunc {
		// func (x receiver) Say(i int) error {...
		methodMeta := strings.SplitN(v, ")", 2)[1]
		methodMeta = strings.SplitN(methodMeta, "{", 2)[0]
		sb.WriteString(methodMeta)
		sb.WriteString(";")
	}
	sb.WriteString("}\n\n")
	// 生成类型名和工厂函数
	fmt.Fprintf(&sb, "type %s struct {\n\t*client.Client\n}\n", typeName)
	fmt.Fprintf(&sb, "func New%s(client *client.Client) %s {", typeName, interName)
	fmt.Fprintf(&sb, "\n\tproxy := &%s{}\n\terr := client.BindFunc(\"%s\",proxy)\n\t", typeName, recvName)
	sb.WriteString("if err != nil {\n\tpanic(err)\n\t}\n\tproxy.Client = client\n\treturn proxy\n}\n")
	return sb.String()
}

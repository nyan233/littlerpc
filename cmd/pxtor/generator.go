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
	"text/template"
	"time"
)

var (
	receiver   = flag.String("r", "", "代理对象的接收器: package.RecvName")
	dir        = flag.String("d", "./", "解析接收器的路径: ./")
	outName    = flag.String("o", "", "输出的文件名，默认的格式: receiver_proxy.go")
	sourceName = flag.String("s", "", "SourceName Example(Hello1.Hello2) SourceName == Hello1")
	genAsync   = flag.Bool("g", false, "生成Async Api, 默认不生成")
	fileSet    *token.FileSet
)

func main() {
	flag.Parse()
	if *receiver == "" {
		panic(interface{}("no receiver specified"))
	}
	if *sourceName == "" {
		*sourceName = *receiver
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
		tmp := getAllFunc(v, rawFile, *sourceName, *genAsync, func(recvT string) bool {
			if recvT == recvName {
				return true
			}
			return false
		})
		funcStrs = append(funcStrs, tmp...)
	}
	fileBuffer.WriteString(createBeforeCode(pkgName, recvName, *sourceName, funcStrs))
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
	file, err := os.OpenFile(*dir+"/"+*outName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		panic(interface{}(err))
	}
	writeN, err := file.Write(fmtBytes)
	if err != nil {
		panic(interface{}(err))
	}
	if writeN != len(fmtBytes) {
		panic(interface{}(errors.New("write format bytes no equal")))
	}
}

func getAllFunc(file *ast.File, rawFile *os.File, sourceName string, genAsync bool, filter func(recvT string) bool) []string {
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
		syncApi, err := genSyncApi(recvName, sourceName, funName, inNameList, inTypeList, outTypeList)
		if err != nil {
			return nil
		}
		funcStrs = append(funcStrs, syncApi)
		if genAsync {
			asyncApi, err := genAsyncApi(recvName, sourceName, funName, inNameList, inTypeList, outTypeList)
			if err != nil {
				return nil
			}
			funcStrs = append(funcStrs, asyncApi[0])
			funcStrs = append(funcStrs, asyncApi[1])
		}
	}
	return funcStrs
}

// 生成同步调用的Api
func genSyncApi(recvName, source, service string, inNameList, inTypeList, outList []string) (string, error) {
	if len(inNameList) != len(inTypeList) {
		return "", errors.New("inNameList and inTypeList length not equal")
	}
	recvName = GetTypeName(recvName)
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "func (p %s) %s(", recvName, service)
	for i := 0; i < len(inNameList); i++ {
		_, _ = fmt.Fprintf(&sb, "%s %s,", inNameList[i], inTypeList[i])
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
		_, _ = fmt.Fprintf(&sb, "{rep,err := p.Call(\"%s.%s\",", source, service)
	} else {
		_, _ = fmt.Fprintf(&sb, "{_,err := p.Call(\"%s.%s\",", source, service)
	}
	for _, v := range inNameList {
		sb.WriteString(v)
		sb.WriteString(",")
	}
	sb.WriteString(");")
	for k, v := range outList[:len(outList)-1] {
		_, _ = fmt.Fprintf(&sb, "r%d,_ := rep[%d].(%s);", k, k, v)
	}
	// 生成最终返回的代码
	sb.WriteString("return ")
	for k := range outList[:len(outList)-1] {
		if k == len(outList)-1 {
			_, _ = fmt.Fprintf(&sb, "r%d", k)
			continue
		}
		_, _ = fmt.Fprintf(&sb, "r%d,", k)
	}
	sb.WriteString("err;}")
	return sb.String(), nil
}

// 生成异步调用的Api
func genAsyncApi(recvName, source, service string, inNameList, inTypeList, outList []string) (asyncApi [2]string, err error) {
	if len(inNameList) != len(inTypeList) {
		return [2]string{}, errors.New("inNameList and inTypeList length not equal")
	}
	recvName = GetTypeName(recvName)
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "func (p %s) Async%s(", recvName, service)
	for i := 0; i < len(inNameList); i++ {
		_, _ = fmt.Fprintf(&sb, "%s %s,", inNameList[i], inTypeList[i])
	}
	_, _ = fmt.Fprintf(&sb, ") error {return p.SyncCall(\"%s.%s\",", source, service)
	for _, v := range inNameList {
		sb.WriteString(v)
		sb.WriteByte(',')
	}
	sb.WriteString(")}")
	asyncApi[0] = sb.String()
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "func (p %sProxy) Register%sCallBack(fn func(", recvName, service)
	for k, v := range outList {
		_, _ = fmt.Fprintf(&sb, "r%s %s,", strconv.Itoa(k), v)
	}
	sb.WriteString("))")
	_, _ = fmt.Fprintf(&sb, "{p.RegisterCallBack(\"%s.%s\",func(rep []interface{}, err error) {", recvName, service)
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
			_, _ = fmt.Fprintf(&sb, "r%d,_ := rep[%d].(%s);", k, k, v)
			continue
		}
		_, _ = fmt.Fprintf(&sb, "r%d := rep[%d].(%s);", k, k, v)
	}
	// 最后生成调用的代码
	sb.WriteString("fn(")
	for k := range outList {
		_, _ = fmt.Fprintf(&sb, "r%d,", k)
	}
	sb.WriteString(");})}")
	asyncApi[1] = sb.String()
	return
}

type BeforeCodeDesc struct {
	PackageName   string
	GeneratorName string
	CreateTime    time.Time
	Author        string
	ImportList    []string
	InterfaceName string
	MethodList    []string
	SourceName    string
	TypeName      string
	RealTypeName  string
}

// 在这里生成包注释、导入、工厂函数、各种需要的类型
func createBeforeCode(pkgName, recvName, source string, allFunc []string) string {
	interfaceName := recvName + "Proxy"
	typeName := GetTypeName(recvName)
	t, err := template.New("BeforeCodeDesc").Parse(BeforeCodeTemplate)
	if err != nil {
		panic(err)
	}
	var sb strings.Builder
	sb.Grow(1024)
	desc := &BeforeCodeDesc{
		PackageName:   pkgName,
		GeneratorName: "littlerpc-generator",
		CreateTime:    time.Now(),
		Author:        "NoAuthor",
		ImportList: []string{
			"github.com/nyan233/littlerpc/client",
		},
		InterfaceName: interfaceName,
		SourceName:    source,
		TypeName:      typeName,
		RealTypeName:  recvName,
	}
	for _, v := range allFunc {
		// func (x receiver) Say(i int) error {...
		methodMeta := strings.SplitN(v, ")", 2)[1]
		methodMeta = strings.SplitN(methodMeta, "{", 2)[0]
		desc.MethodList = append(desc.MethodList, methodMeta)
	}
	err = t.Execute(&sb, desc)
	if err != nil {
		panic(err)
	}
	return sb.String()
}

func GetTypeName(recvName string) string {
	if len(recvName) == 0 {
		return ""
	}
	bytes4Str := []byte(recvName)
	lowBytes := bytes.ToLower(bytes4Str[:1])
	bytes4Str[0] = lowBytes[0]
	return string(bytes4Str) + "Impl"
}

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	flag "github.com/spf13/pflag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type GenMethod func(receive Argument, name, service string, input []Argument, output []Argument) (getResult func() string, err error)

const (
	SyncStyle     = "sync"
	AsyncStyle    = "async"
	RequestsStyle = "requests"
)

var (
	receiver   = flag.StringP("receive", "r", "", "代理对象的接收器: package.RecvName")
	dir        = flag.StringP("dir", "d", "./", "解析接收器的路径: ./")
	outName    = flag.StringP("out", "o", "", "输出的文件名，默认的格式: receiver_proxy.go")
	sourceName = flag.StringP("source", "s", "", "SourceName Example(Hello1.Hello2) SourceName == Hello1")
	generateId = flag.BoolP("gen_id", "i", false, "生成唯一id, 多个文件在同一个包时binder/caller不会冲突, 但对于mock场景不友好")
	// TODO: 实现不同API风格的生成函数
	style   = flag.StringP("gen", "g", SyncStyle, "生成的API风格, TODO")
	fileSet *token.FileSet
)

func main() {
	flag.Parse()
	if *receiver == "" {
		panic(interface{}("no receiver specified"))
	}
	if *sourceName == "" {
		*sourceName = strings.Split(*receiver, ".")[1]
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
	var genFn GenMethod
	switch *style {
	case SyncStyle:
		genFn = genSync
	default:
		panic("no support gen style")
	}
	usageImportNameAndPath := make(map[string]string)
	ignoreSetup := ignoreSetup(pkgDir.Files, *receiver)
	for k, v := range pkgDir.Files {
		rawFile, err := os.Open(path.Dir(*dir) + "/" + k)
		if err != nil {
			panic(interface{}(err))
		}
		tmp := getAllFunc(v, rawFile, usageImportNameAndPath, *sourceName, genFn, func(recvT string) bool {
			if recvT == recvName {
				return true
			}
			return false
		}, ignoreSetup)
		funcStrs = append(funcStrs, tmp...)
	}
	fileBuffer.WriteString(createBeforeCode(pkgName, recvName, *sourceName, funcStrs, usageImportNameAndPath))
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

func getAllFunc(file *ast.File, rawFile *os.File, usageImportNameAndPath map[string]string, sourceName string,
	genFunc GenMethod, filter func(recvT string) bool, ignoreSetup bool) []string {
	funcStrs := make([]string, 0)
	importNamePathMapping := buildImportNameAndPath(file.Imports)
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
		if funName == "Setup" && ignoreSetup {
			continue
		}
		inputList := make([]Argument, 0, 4)
		outputList := make([]Argument, 0, 4)
		// 处理参数的序列化
		for _, pv := range funcDecl.Type.Params.List {
			for _, pvName := range pv.Names {
				arg := Argument{
					Name: pvName.Name,
					Type: handleAstType(pv.Type, rawFile),
				}
				inputList = append(inputList, arg)
				// usage import ?
				if !strings.Contains(arg.Type, ".") {
					continue
				}
				typeName := strings.Trim(arg.Type, "*")
				importName := strings.SplitN(typeName, ".", 2)[0]
				usageImportNameAndPath[importName] = importNamePathMapping[importName]
			}
		}
		// 找出所有的返回值类型
		for _, rv := range funcDecl.Type.Results.List {
			res := Argument{
				Type: handleAstType(rv.Type, rawFile),
			}
			outputList = append(outputList, res)
			// usage import ?
			if !strings.Contains(res.Type, ".") {
				continue
			}
			typeName := strings.Trim(res.Type, "*")
			importName := strings.SplitN(typeName, ".", 2)[0]
			usageImportNameAndPath[importName] = importNamePathMapping[importName]
		}
		after, err := genFunc(Argument{
			Name: "p",
			Type: recvName,
		}, funName, sourceName+"."+funName, inputList, outputList)
		if err != nil {
			return nil
		}
		funcStrs = append(funcStrs, after())
	}
	return funcStrs
}

func ignoreSetup(astFiles map[string]*ast.File, receive string) (ignore bool) {
	for _, astFile := range astFiles {
		importNameAndPath := buildImportNameAndPath(astFile.Imports)
		for _, decl := range astFile.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
			if !ok {
				continue
			}
			targetTypeName := strings.Split(receive, ".")[1]
			if typeSpec.Name.Name != targetTypeName {
				continue
			}
			// 只有Struct类型才能内嵌RpcServer
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structType.Fields.List {
				se, ok := field.Type.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				if se.Sel.Name != "RpcServer" {
					continue
				}
				ident, ok := se.X.(*ast.Ident)
				if !ok {
					continue
				}
				importPath := importNameAndPath[ident.Name]
				if importPath != "github.com/nyan233/littlerpc/core/server" {
					continue
				}
				return true
			}
		}
	}
	return
}

func buildImportNameAndPath(imports []*ast.ImportSpec) map[string]string {
	result := make(map[string]string, len(imports))
	for _, v := range imports {
		// 没有别名
		pathVal := strings.Trim(v.Path.Value, "\"")
		if v.Name == nil {
			tmp := strings.Split(pathVal, "/")
			result[tmp[len(tmp)-1]] = pathVal
			continue
		}
		result[v.Name.Name] = pathVal
	}
	return result
}

// 生成同步调用的Api
func genSync(receive Argument, name, service string, input []Argument, output []Argument) (getResult func() string, err error) {
	receive.Type = GetTypeName(receive.Type)
	m := Method{
		Receive:     receive,
		ServiceName: service,
		Name:        name,
		InputList:   input,
		OutputList:  output,
		Statement:   Statement{},
	}
	return m.FormatToSync, nil
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

type ImportDesc struct {
	Name string
	Path string
}

type BeforeCodeDesc struct {
	PackageName   string
	GeneratorName string
	CreateTime    time.Time
	Author        string
	ImportList    []ImportDesc
	InterfaceName string
	MethodList    []string
	SourceName    string
	TypeName      string
	RealTypeName  string
	GenId         string
}

// 在这里生成包注释、导入、工厂函数、各种需要的类型
func createBeforeCode(pkgName, recvName, source string, allFunc []string, usageImportNameAndPath map[string]string) string {
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
		GeneratorName: "pxtor",
		CreateTime:    time.Now(),
		Author:        "NoAuthor",
		ImportList: []ImportDesc{
			{
				Path: "github.com/nyan233/littlerpc/core/client",
			},
		},
		InterfaceName: interfaceName,
		SourceName:    source,
		TypeName:      typeName,
		RealTypeName:  recvName,
	}
	for importName, importPath := range usageImportNameAndPath {
		// 未使用别名
		if strings.HasSuffix(importPath, importName) {
			desc.ImportList = append(desc.ImportList, ImportDesc{Path: importPath})
			continue
		}
		desc.ImportList = append(desc.ImportList, ImportDesc{importName, importPath})
	}
	if *generateId {
		desc.GenId = getId()
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

func getId() string {
	after := time.Now().UnixNano()
	rand.Seed(after)
	before := rand.Uint64()
	bStr := hex.EncodeToString(binary.BigEndian.AppendUint64(nil, before))
	aStr := hex.EncodeToString(binary.BigEndian.AppendUint64(nil, uint64(after)))
	return aStr + bStr
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

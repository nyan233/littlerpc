package server

import (
	"context"
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"github.com/nyan233/littlerpc/core/common/metadata"
	"github.com/nyan233/littlerpc/core/middle/codec"
	"github.com/nyan233/littlerpc/core/middle/packer"
	"reflect"
	_ "unsafe"
)

var (
	//go:linkname codecViewer github.com/nyan233/littlerpc/core/middle/codec.codecCollection
	codecViewer map[string]codec.Codec
	//go:linkname packerViewer github.com/nyan233/littlerpc/core/middle/packer.packerCollection
	packerViewer map[string]packer.Packer
)

type MethodTable struct {
	SourceName string
	Table      []string
}

// ArgumentType 参数类型 == map[string]interface{}表示struct,interface{}里的参数也同理 == string时表示其它类型
type ArgumentType struct {
	Kind string      `json:"kind"`
	Type interface{} `json:"type"`
}

// LittleRpcReflection 反射服务
type LittleRpcReflection struct {
	rpcServer *Server
}

func (l *LittleRpcReflection) MethodTable(ctx context.Context, sourceName string) (*MethodTable, error) {
	source, ok := l.rpcServer.sources.LoadOk(sourceName)
	if !ok {
		return nil, errorhandler.ServiceNotfound
	}
	mt := &MethodTable{
		SourceName: sourceName,
		Table:      nil,
	}
	for methodName := range source.ProcessSet {
		mt.Table = append(mt.Table, methodName)
	}
	return mt, nil
}

func (l *LittleRpcReflection) AllInstance(ctx context.Context) (map[string]string, error) {
	mp := make(map[string]string, 4)
	l.rpcServer.sources.Range(func(key string, value *metadata.Source) bool {
		mp[key] = value.InstanceType.Name()
		return true
	})
	return mp, nil
}

func (l *LittleRpcReflection) AllCodec(ctx context.Context) ([]string, error) {
	codecSchemeSet := make([]string, 0, len(codecViewer))
	for k := range codecViewer {
		codecSchemeSet = append(codecSchemeSet, k)
	}
	return codecSchemeSet, nil
}

func (l *LittleRpcReflection) AllPacker(ctx context.Context) ([]string, error) {
	packerSchemeSet := make([]string, 0, len(packerViewer))
	for k := range packerViewer {
		packerSchemeSet = append(packerSchemeSet, k)
	}
	return packerSchemeSet, nil
}

func (l *LittleRpcReflection) MethodArgumentType(ctx context.Context, serviceName string) ([]ArgumentType, error) {
	service, ok := l.rpcServer.services.LoadOk(serviceName)
	if !ok {
		return nil, errorhandler.ServiceNotfound
	}
	argDesc := make([]ArgumentType, 0, 4)
	typ := service.Value.Type()
	for i := 0; i < typ.NumIn(); i++ {
		in := typ.In(i)
		argDesc = append(argDesc, getArgumentType(in))
	}
	return argDesc, nil
}

func getArgumentType(typ reflect.Type) ArgumentType {
	switch typ.Kind() {
	case reflect.Struct:
		fieldMap := make(map[string]interface{}, typ.NumField())
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldMap[field.Name] = getArgumentType(field.Type)
		}
		return ArgumentType{
			Kind: "struct",
			Type: fieldMap,
		}
	case reflect.Interface:
		return ArgumentType{
			Kind: "interface",
			Type: typ.PkgPath() + "." + typ.Name(),
		}
	case reflect.Slice:
		return ArgumentType{
			Kind: "array",
			Type: getTypeName(typ),
		}
	case reflect.Map:
		return ArgumentType{
			Kind: "map",
			Type: struct {
				Key   string
				Value string
			}{
				Key:   getTypeName(typ.Key()),
				Value: getTypeName(typ.Elem()),
			},
		}
	case reflect.String:
		return ArgumentType{
			Kind: "string",
			Type: typ.Name(),
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return ArgumentType{
			Kind: "integer",
			Type: typ.Name(),
		}
	case reflect.Pointer:
		return getArgumentType(typ.Elem())
	default:
		return ArgumentType{
			Kind: "unknown",
			Type: nil,
		}
	}
}

func getTypeName(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Slice:
		return "[]" + getTypeName(typ.Elem())
	case reflect.Pointer:
		return "*" + getTypeName(typ.Elem())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Interface:
		return typ.PkgPath() + "." + typ.Name()
	case reflect.Map:
		return "Map[" + getTypeName(typ.Key()) + "," + getTypeName(typ.Elem()) + "]"
	default:
		return typ.Name()
	}
}

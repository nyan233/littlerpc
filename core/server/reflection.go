package server

import (
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

type ArgumentType struct {
	TypeName string      `json:"type_name"`
	Other    interface{} `json:"other"`
}

// LittleRpcReflection 反射服务
type LittleRpcReflection struct {
	rpcServer *Server
}

func (l *LittleRpcReflection) MethodTable(sourceName string) (*MethodTable, error) {
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

func (l *LittleRpcReflection) AllInstance() (map[string]string, error) {
	mp := make(map[string]string, 4)
	l.rpcServer.sources.Range(func(key string, value *metadata.Source) bool {
		mp[key] = value.InstanceType.Name()
		return true
	})
	return mp, nil
}

func (l *LittleRpcReflection) AllCodec() ([]string, error) {
	codecSchemeSet := make([]string, 0, len(codecViewer))
	for k := range codecViewer {
		codecSchemeSet = append(codecSchemeSet, k)
	}
	return codecSchemeSet, nil
}

func (l *LittleRpcReflection) AllPacker() ([]string, error) {
	packerSchemeSet := make([]string, 0, len(packerViewer))
	for k := range packerViewer {
		packerSchemeSet = append(packerSchemeSet, k)
	}
	return packerSchemeSet, nil
}

func (l *LittleRpcReflection) MethodArgumentType(serviceName string) ([]*ArgumentType, error) {
	service, ok := l.rpcServer.services.LoadOk(serviceName)
	if !ok {
		return nil, errorhandler.ServiceNotfound
	}
	argDesc := make([]*ArgumentType, 0, 4)
	typ := service.Value.Type()
	for i := 1; i < typ.NumIn(); i++ {
		in := typ.In(i)
		typDesc := &ArgumentType{}
		switch in.Kind() {
		case reflect.Slice:
			break
		case reflect.Map:
			break
		case reflect.Struct:
			break
		default:
			typDesc.TypeName = in.Name()
		}
		argDesc = append(argDesc, typDesc)
	}
	return argDesc, nil
}

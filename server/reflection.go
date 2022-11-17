package server

import (
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"reflect"
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
		return nil, common.ServiceNotfound
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

func (l *LittleRpcReflection) MethodArgumentType(serviceName string) ([]*ArgumentType, error) {
	service, ok := l.rpcServer.services.LoadOk(serviceName)
	if !ok {
		return nil, common.ServiceNotfound
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

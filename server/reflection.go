package server

import (
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/pkg/container"
	"reflect"
)

type MethodTable struct {
	InstanceName string
	Table        []string
}

type ArgumentType struct {
	TypeName string      `json:"type_name"`
	Other    interface{} `json:"other"`
}

// LittleRpcReflection 反射服务
type LittleRpcReflection struct {
	elems *container.SyncMap118[string, metadata.ElemMeta]
}

func (l *LittleRpcReflection) MethodTable(instanceName string) (*MethodTable, error) {
	elem, ok := l.elems.LoadOk(instanceName)
	if !ok {
		return nil, common.ErrMethodNoRegister
	}
	mt := &MethodTable{
		InstanceName: instanceName,
		Table:        nil,
	}
	for methodName := range elem.Methods {
		mt.Table = append(mt.Table, methodName)
	}
	return mt, nil
}

func (l *LittleRpcReflection) AllInstance() (map[string]string, error) {
	mp := make(map[string]string, 4)
	l.elems.Range(func(key string, value metadata.ElemMeta) bool {
		mp[key] = value.Typ.String()
		return true
	})
	return mp, nil
}

func (l *LittleRpcReflection) MethodArgumentType(instanceName, methodName string) ([]*ArgumentType, error) {
	elem, ok := l.elems.LoadOk(instanceName)
	if !ok {
		return nil, common.ErrMethodNoRegister
	}
	method, ok := elem.Methods[methodName]
	if !ok {
		return nil, common.ErrMethodNoRegister
	}
	argDesc := make([]*ArgumentType, 0, 4)
	typ := method.Value.Type()
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

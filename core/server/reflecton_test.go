package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nyan233/littlerpc/core/common/metadata"
	"reflect"
	"testing"

	"github.com/nyan233/littlerpc/core/container"
)

func TestReflection(t *testing.T) {
	noneServer := &Server{
		services: *container.NewRCUMap[string, *metadata.Process](),
		sources:  *container.NewRCUMap[string, *metadata.Source](),
	}
	reflection := &LittleRpcReflection{
		rpcServer: noneServer,
	}
	err := noneServer.RegisterClass(ReflectionSource, reflection, nil)
	if err != nil {
		t.Fatal(err)
	}
	table, err := reflection.MethodTable(context.Background(), ReflectionSource)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(*table)
	allInstance, err := reflection.AllInstance(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(allInstance)
	argumentType, err := reflection.MethodArgumentType(context.Background(), fmt.Sprintf("%s.%s", ReflectionSource, "MethodArgumentType"))
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.Marshal(argumentType)
	t.Log(string(bytes))
	allCodec, err := reflection.AllCodec(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ = json.Marshal(allCodec)
	t.Log(string(bytes))
	allPacker, err := reflection.AllPacker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ = json.Marshal(allPacker)
	t.Log(string(bytes))
}

func TestGetArgumentType(t *testing.T) {
	type FuncArg1 struct {
		Uid     int64
		Name    string
		Comment string
		Cancel  context.Context
		Friend  []FuncArg1
	}
	type Func1 func(ctx context.Context, a1 *FuncArg1, a2 string, a3 map[string]map[uint64]FuncArg1)
	typ := reflect.TypeOf((Func1)(nil))
	for i := 0; i < typ.NumIn(); i++ {
		bytes, err := json.Marshal(getArgumentType(typ.In(i)))
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(bytes))
	}
}

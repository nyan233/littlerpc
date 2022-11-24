package server

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/pkg/container"
)

func TestReflection(t *testing.T) {
	noneServer := &Server{
		services: container.SyncMap118[string, *metadata.Process]{},
	}
	reflection := &LittleRpcReflection{
		rpcServer: noneServer,
	}
	err := noneServer.RegisterClass(ReflectionSource, reflection, nil)
	if err != nil {
		t.Fatal(err)
	}
	table, err := reflection.MethodTable(ReflectionSource)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(*table)
	allInstance, err := reflection.AllInstance()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(allInstance)
	argumentType, err := reflection.MethodArgumentType(fmt.Sprintf("%s.%s", ReflectionSource, "MethodArgumentType"))
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.Marshal(argumentType)
	t.Log(string(bytes))
	allCodec, err := reflection.AllCodec()
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ = json.Marshal(allCodec)
	t.Log(string(bytes))
	allPacker, err := reflection.AllPacker()
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ = json.Marshal(allPacker)
	t.Log(string(bytes))
}

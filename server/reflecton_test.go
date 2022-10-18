package server

import (
	"encoding/json"
	"github.com/nyan233/littlerpc/pkg/common"
	"github.com/nyan233/littlerpc/pkg/container"
	"testing"
)

func TestReflection(t *testing.T) {
	noneServer := &Server{
		elems: container.SyncMap118[string, common.ElemMeta]{},
	}
	reflection := &LittleRpcReflection{
		elems: &noneServer.elems,
	}
	err := noneServer.RegisterClass(reflection, nil)
	if err != nil {
		t.Fatal(err)
	}
	table, err := reflection.MethodTable("LittleRpcReflection")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(*table)
	allInstance, err := reflection.AllInstance()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(allInstance)
	argumentType, err := reflection.MethodArgumentType("LittleRpcReflection", "MethodArgumentType")
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.Marshal(argumentType)
	t.Log(string(bytes))
}

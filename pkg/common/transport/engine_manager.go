package transport

import "github.com/nyan233/littlerpc/pkg/container"

var (
	EngineManager = &engineManager{}
)

type engineManager struct {
	serverEngineCollection container.SyncMap118[string, NewServerEngineBuilder]
	clientEngineCollection container.SyncMap118[string, NewClientEngineBuilder]
}

func (m *engineManager) RegisterServerEngine(scheme string, builder NewServerEngineBuilder) {
	m.serverEngineCollection.Store(scheme, builder)
}

func (m *engineManager) GetServerEngine(scheme string) NewServerEngineBuilder {
	builder, _ := m.serverEngineCollection.LoadOk(scheme)
	return builder
}

func (m *engineManager) RegisterClientEngine(scheme string, builder NewClientEngineBuilder) {
	m.clientEngineCollection.Store(scheme, builder)
}

func (m *engineManager) GetClientEngine(scheme string) NewClientEngineBuilder {
	builder, _ := m.clientEngineCollection.LoadOk(scheme)
	return builder
}

func init() {
	EngineManager.RegisterServerEngine("nbio_tcp", NewNBioTcpServerEngine)
	EngineManager.RegisterClientEngine("nbio_tcp", NewNBioTcpClientEngine)
	EngineManager.RegisterServerEngine("nbio_ws", NewNBioWebsocketServerEngine)
	EngineManager.RegisterClientEngine("nbio_ws", NewNBioWebsocketClientEngine)
	EngineManager.RegisterServerEngine("std_tcp", NewStdTcpServerEngine)
	EngineManager.RegisterClientEngine("std_tcp", NewStdTcpClientEngine)
}

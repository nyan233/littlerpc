package transport

import "github.com/nyan233/littlerpc/core/container"

var (
	Manager = &manager{}
)

type manager struct {
	serverEngineCollection container.SyncMap118[string, NewServerBuilder]
	clientEngineCollection container.SyncMap118[string, NewClientBuilder]
}

func (m *manager) RegisterServerEngine(scheme string, builder NewServerBuilder) {
	m.serverEngineCollection.Store(scheme, builder)
}

func (m *manager) GetServerEngine(scheme string) NewServerBuilder {
	builder, _ := m.serverEngineCollection.LoadOk(scheme)
	return builder
}

func (m *manager) RegisterClientEngine(scheme string, builder NewClientBuilder) {
	m.clientEngineCollection.Store(scheme, builder)
}

func (m *manager) GetClientEngine(scheme string) NewClientBuilder {
	builder, _ := m.clientEngineCollection.LoadOk(scheme)
	return builder
}

func init() {
	Manager.RegisterServerEngine("nbio_tcp", NewNBioTcpServer)
	Manager.RegisterClientEngine("nbio_tcp", NewNBioTcpClient)
	Manager.RegisterServerEngine("nbio_ws", NewNBioWebsocketServer)
	Manager.RegisterClientEngine("nbio_ws", NewNBioWebsocketClient)
	Manager.RegisterServerEngine("std_tcp", NewStdTcpServer)
	Manager.RegisterClientEngine("std_tcp", NewStdTcpClient)
}

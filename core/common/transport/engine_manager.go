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
	Manager.RegisterServerEngine("nbio_tcp", func(config NetworkServerConfig) ServerBuilder {
		return NewNBioBaseServer("tcp", config)
	})
	Manager.RegisterClientEngine("nbio_tcp", func() ClientBuilder {
		return NewNBioBaseClient("tcp")
	})
	Manager.RegisterServerEngine("nbio_udp", func(config NetworkServerConfig) ServerBuilder {
		return NewNBioBaseServer("udp", config)
	})
	Manager.RegisterClientEngine("nbio_udp", func() ClientBuilder {
		return NewNBioBaseClient("udp")
	})
	Manager.RegisterServerEngine("nbio_unix", func(config NetworkServerConfig) ServerBuilder {
		return NewNBioBaseServer("unix", config)
	})
	Manager.RegisterClientEngine("nbio_unix", func() ClientBuilder {
		return NewNBioBaseClient("unix")
	})
	Manager.RegisterServerEngine("nbio_ws", NewNBioWebsocketServer)
	Manager.RegisterClientEngine("nbio_ws", NewNBioWebsocketClient)
	Manager.RegisterServerEngine("std_tcp", func(config NetworkServerConfig) ServerBuilder {
		return NewStdNetServer("tcp", config)
	})
	Manager.RegisterClientEngine("std_tcp", func() ClientBuilder {
		return NewStdNetClient("tcp")
	})
	Manager.RegisterServerEngine("std_udp", func(config NetworkServerConfig) ServerBuilder {
		return NewStdNetServer("udp", config)
	})
	Manager.RegisterClientEngine("std_udp", func() ClientBuilder {
		return NewStdNetClient("udp")
	})
	Manager.RegisterServerEngine("std_unix", func(config NetworkServerConfig) ServerBuilder {
		return NewStdNetServer("unix", config)
	})
	Manager.RegisterClientEngine("std_unix", func() ClientBuilder {
		return NewStdNetClient("unix")
	})
}

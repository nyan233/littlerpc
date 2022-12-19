package transport

import "net"

type NetworkClientConfig struct {
	ServerAddr string
	KeepAlive  bool
	TLSPubPem  []byte
	TLSPriPem  []byte
	Dialer     *net.Dialer
}

type NetworkServerConfig struct {
	Addrs     []string
	KeepAlive bool
	TLSPubPem []byte
}

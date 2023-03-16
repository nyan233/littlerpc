package transport

type NetworkClientConfig struct {
	ServerAddr string
	KeepAlive  bool
	TLSPubPem  []byte
	TLSPriPem  []byte
}

type NetworkServerConfig struct {
	Addrs     []string
	KeepAlive bool
	TLSPubPem []byte
}

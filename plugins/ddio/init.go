package ddio

import "github.com/nyan233/littlerpc/core/common/transport"

func init() {
	transport.Manager.RegisterServerEngine("ddio_tcp", nil)
	transport.Manager.RegisterClientEngine("ddio_tcp", nil)
}

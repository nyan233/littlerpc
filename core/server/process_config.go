package server

import (
	"github.com/nyan233/littlerpc/core/common/errorhandler"
	"strings"
	"unsafe"
)

type RpcServer struct {
	Prefix string
	s      *Server
}

func (r *RpcServer) init(prefix string, s *Server) {
	r.Prefix = prefix
	r.s = s
}

func (r *RpcServer) Setup() {
	return
}

func (r *RpcServer) HijackProcess(name string, fn func(stub *Stub)) error {
	pName := strings.Join([]string{r.Prefix, name}, ".")
	process, ok := r.s.services.LoadOk(pName)
	if !ok {
		return errorhandler.ServiceNotfound
	}
	process.Hijack = true
	process.Hijacker = unsafe.Pointer(&fn)
	return nil
}

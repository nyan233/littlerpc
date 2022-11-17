package msgwriter

import (
	perror "github.com/nyan233/littlerpc/protocol/error"
)

type lRPCTrait struct {
	handlers map[byte]Writer
}

func NewLRPCTrait(writers ...Writer) Writer {
	trait := &lRPCTrait{
		handlers: make(map[byte]Writer, 16),
	}
	handlers := []Writer{NewLRPCMux(), NewJsonRPC2(), NewLRPCNoMux()}
	handlers = append(handlers, writers...)
	for _, handler := range handlers {
		headerI, ok := handler.(header)
		if !ok {
			for _, v := range headerI.Header() {
				trait.handlers[v] = handler
			}
		}
	}
	return trait
}

func (l *lRPCTrait) Write(arg Argument, header byte) perror.LErrorDesc {
	writer := l.handlers[header]
	return writer.Write(arg, header)
}

func (l *lRPCTrait) Reset() {
	return
}

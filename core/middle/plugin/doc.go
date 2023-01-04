// Package plugin is plugin interface
//
//	Plugin Context never private, the context is the request internal global shared
//	able usage private type for access self data, server plugin example:
//	type xxPluginService struct {}
//
//	-->HandleReceive(...,kvAppend,...)
//	-->kvAppend(xxPluginService{},"hello world")
//
//	---->HandleCall(pCtx,...)
//	---->value,_ := pCtx.Value(xxPluginService{}).(string)
//	---->value == "hello world"
package plugin

type InjectServerPlugin func(name string, plugin ServerPlugin)

type InjectClientPlugin func(name string, plugin ClientPlugin)

package error

/*
	LittleRpc关于Error的设计很大一部份源自grpc,grpc Error的设计确实是一个好想法
	See : https://github.com/grpc/grpc-go/blob/master/internal/status/status.go
*/

/*
	本包提供了Little-Rpc中传递自定义错误的能力, LStdError为默认使用的实例, 用户可以通过自己实现LNewErrorDesc工厂
	来接入自己的逻辑, 用户过程返回的错误是内置的go error时, Code == Unknown , 所有的client&server必须这么实现
*/

package jsonrpc2

import "encoding/json"

const (
	Header       = '{'
	KeepAlive    = "rpc.keepalive"
	RequestType  = "request"
	ResponseType = "response"
	Version      = "2.0"
)

const (
	ErrorParser    = -32700 // jsonrpc2 解析消息失败
	InvalidRequest = -32600 // 无效的请求
	MethodNotFound = -32601 // 找不到方法
	InvalidParams  = -32602 // 无效的参数
	ErrorInternal  = -32603 // 内部错误
	Unknown        = -32004 // 未知的错误
)

type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type Response struct {
	BaseMessage
	ResponseTrait
}

type ResponseTrait struct {
	Result [][]byte `json:"result"`
	Error  *Error   `json:"error,omitempty"`
}

type Request struct {
	BaseMessage
	RequestTrait
}

type RequestTrait struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type BaseMessage struct {
	MessageType int               `json:"rpc_message_type"`
	Version     string            `json:"jsonrpc"`
	Id          uint64            `json:"id"`
	MetaData    map[string]string `json:"rpc_metadata"`
}

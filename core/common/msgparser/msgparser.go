package msgparser

import (
	"github.com/nyan233/littlerpc/core/common/inters"
	"github.com/nyan233/littlerpc/core/protocol/message"
)

const (
	_ScanInit      int = iota // 初始化状态, 扫描到数据包的第1个Byte时
	_ScanMsgParse1            // 扫描基本信息状态, 扫描到描述数据包的基本信息, 这些信息可以确定数据包的长度/MsgId
	_ScanMsgParse2            // 扫描完整数据状态, 扫描完整数据包
)

const (
	DefaultBufferSize = 4096    // 4KB
	MaxBufferSize     = 1 << 20 // 1MB
	DefaultParser     = "lrpc-trait"
)

type ParserMessage struct {
	Message *message.Message
	// 没有Header特征的协议可以选定一个固定的虚拟值, 由Parser返回
	// 再将这个固定的虚拟值注册到Writer中以找到对应的Writer
	Header byte
}

type Factory func(msgAllocator AllocTor, bufSize uint32) Parser

// Parser 解析器的所有接口的实现必须是线程安全/goroutine safe
// 否则则会出现data race/race conditions
type Parser interface {
	// Parse 处理数据的接口必须能够正确处理half-package
	// 也必须能处理有多个完整报文的数据, 在解析失败时返回对应的error
	Parse(data []byte) (msgs []ParserMessage, err error)
	// Free 用于释放Parse返回的数据, 在Parse返回error时这个过程
	// 绝对不能被调用
	Free(msg *message.Message)
	inters.Reset
}

var (
	parserFactoryCollections = make(map[string]Factory)
)

func Register(scheme string, pf Factory) {
	if pf == nil {
		panic("parser factory is nil")
	}
	if scheme == "" {
		panic("parser scheme is empty")
	}
	parserFactoryCollections[scheme] = pf
}

func Get(scheme string) Factory {
	return parserFactoryCollections[scheme]
}

func init() {
	Register(DefaultParser, NewLRPCTrait)
}

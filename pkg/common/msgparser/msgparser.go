package msgparser

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/pkg/utils"
	"github.com/nyan233/littlerpc/protocol"
	"sync"
)

const (
	_ScanInit int = iota
	_ScanMsgParse1
	_ScanMsgParse2
)

type readyBuffer struct {
	MsgId         uint64
	PayloadLength uint32
	RawBytes      []byte
}

func (b *readyBuffer) IsUnmarshal() bool {
	return b.MsgId != 0 && b.PayloadLength != 0
}

// LMessageParser LParser 负责消息的处理
type LMessageParser struct {
	mu sync.Mutex
	// 简单的分配器接口, 用于分配可复用的Message
	allocTor AllocTor
	// 下一个状态的触发间隔, 也就是距离转移到下一个状态需要读取的数据量
	clickInterval int
	// 当前parser选中的handler
	handler MessageHandler
	// 当前在状态机中处于的状态
	state         int
	noReadyBuffer map[uint64]readyBuffer
	// 存储半包数据的缓冲区, 只有在读完了一个完整的payload的消息的数据包
	// 才会被直接提升到noReadyBuffer中, noMux类型的数据包则不会被提升到
	// noReadyBuffer中, 将完整的消息读取完毕后直接触发onComplete
	halfBuffer []byte
}

func NewLMessageParser(allocTor AllocTor) *LMessageParser {
	return &LMessageParser{
		allocTor:      allocTor,
		clickInterval: 1,
		state:         _ScanInit,
		noReadyBuffer: make(map[uint64]readyBuffer, 16),
		halfBuffer:    make([]byte, 0, protocol.MuxMessageBlockSize),
	}
}

// ParseMsg io.Reader主要用来标识一个读取到半包的连接, 并不会真正去调用他的方法
func (h *LMessageParser) ParseMsg(data []byte) ([]*protocol.Message, error) {
	if h.clickInterval == 1 && len(data) == 0 {
		return nil, errors.New("data length == 0")
	}
	allMsg := make([]*protocol.Message, 0, 4)
	for {
		if len(data) == 0 {
			return allMsg, nil
		}
		switch h.state {
		case _ScanInit:
			h.halfBuffer = append(h.halfBuffer, data[0])
			data = data[1:]
			if handler := GetMessageHandler(h.halfBuffer[0]); handler != nil {
				h.handler = handler
			} else {
				return nil, errors.New(fmt.Sprintf("MagicNumber no MessageHandler -> %d", data[0]))
			}
			h.state = _ScanMsgParse1
			h.clickInterval = h.handler.BaseLen() - 1
		case _ScanMsgParse1:
			readN, readData := utils.ReadFromData(h.clickInterval, data)
			h.halfBuffer = append(h.halfBuffer, readData...)
			data = data[readN:]
			if readN < h.clickInterval {
				h.clickInterval -= readN
				continue
			}
			h.clickInterval = h.handler.MessageLength(h.halfBuffer) - h.handler.BaseLen()
			h.state = _ScanMsgParse2
		case _ScanMsgParse2:
			readN, readData := utils.ReadFromData(h.clickInterval, data)
			h.halfBuffer = append(h.halfBuffer, readData...)
			data = data[readN:]
			if readN < h.clickInterval {
				h.clickInterval -= readN
				continue
			}
			msg := h.allocTor.AllocMessage()
			msg.Reset()
			action, err := h.handler.Unmarshal(h.halfBuffer, msg)
			if err != nil {
				return nil, err
			}
			switch action {
			case UnmarshalBase:
				buf, ok := h.noReadyBuffer[msg.MsgId]
				if !ok {
					h.noReadyBuffer[msg.MsgId] = readyBuffer{
						MsgId:         msg.MsgId,
						PayloadLength: msg.PayloadLength,
						RawBytes:      h.halfBuffer,
					}
				}
				buf.RawBytes = append(buf.RawBytes, readData...)
				if uint32(len(buf.RawBytes)) == buf.PayloadLength {
					msg.Reset()
					err := protocol.UnmarshalMessage(buf.RawBytes, msg)
					if err != nil {
						return nil, err
					}
					allMsg = append(allMsg, msg)
					// 置空/删除Map Key让内存得以回收
					h.noReadyBuffer[msg.MsgId] = readyBuffer{}
					delete(h.noReadyBuffer, msg.MsgId)
				}
			case UnmarshalComplete:
				allMsg = append(allMsg, msg)
			}
			h.clickInterval = 1
			h.state = _ScanInit
			h.halfBuffer = h.halfBuffer[:0]
		}
	}
}

func (h *LMessageParser) FreeMessage(msg *protocol.Message) {
	h.allocTor.FreeMessage(msg)
}

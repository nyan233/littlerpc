package msgparser

import (
	"errors"
	"fmt"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils"
	"sync"

	"github.com/nyan233/littlerpc/internal/reflect"
)

type readyBuffer struct {
	MsgId         uint64
	PayloadLength uint32
	RawBytes      []byte
}

func (b *readyBuffer) IsUnmarshal() bool {
	return b.MsgId != 0 && b.PayloadLength != 0
}

// lRPCTrait 特征化Parser, 根据Header自动选择适合的Handler
type lRPCTrait struct {
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

func NewLRPCTrait(allocTor AllocTor, bufSize uint32) Parser {
	if bufSize > MaxBufferSize {
		bufSize = MaxBufferSize
	} else if bufSize == 0 {
		bufSize = DefaultBufferSize
	}
	return &lRPCTrait{
		allocTor:      allocTor,
		clickInterval: 1,
		state:         _ScanInit,
		noReadyBuffer: make(map[uint64]readyBuffer, 16),
		halfBuffer:    make([]byte, 0, bufSize),
	}
}

// Parse io.Reader主要用来标识一个读取到半包的连接, 并不会真正去调用他的方法
func (h *lRPCTrait) Parse(data []byte) (msgs []ParserMessage, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clickInterval == 1 && len(data) == 0 {
		return nil, errors.New("data length == 0")
	}
	allMsg := make([]ParserMessage, 0, 4)
	defer func() {
		if err != nil {
			h.ResetScan()
			if len(allMsg) == 0 {
				return
			}
			for _, msg := range allMsg {
				h.allocTor.FreeMessage(msg.Message)
			}
		}
	}()
	for {
		if len(data) == 0 {
			return allMsg, nil
		}
		switch h.state {
		case _ScanInit:
			h.halfBuffer = append(h.halfBuffer, data[0])
			data = data[1:]
			if handler := GetHandler(h.halfBuffer[0]); handler != nil {
				h.handler = handler
			} else {
				return nil, errors.New(fmt.Sprintf("MagicNumber no MessageHandler -> %d", data[0]))
			}
			h.state = _ScanMsgParse1
			opt, baseLen := h.handler.BaseLen()
			if opt == SingleRequest {
				msg := h.allocTor.AllocMessage()
				msg.Reset()
				defer func() {
					if err != nil {
						h.allocTor.FreeMessage(msg)
					}
					h.ResetScan()
				}()
				action, err := h.handler.Unmarshal(reflect.SliceBackSpace(data, 1).([]byte), msg)
				if err != nil {
					return nil, err
				}
				err = h.handleAction(action, msg, &allMsg, nil)
				if err != nil {
					return nil, err
				}
				return allMsg, nil
			}
			h.clickInterval = baseLen - 1
		case _ScanMsgParse1:
			readN, readData := utils.ReadFromData(h.clickInterval, data)
			h.halfBuffer = append(h.halfBuffer, readData...)
			data = data[readN:]
			if readN < h.clickInterval {
				h.clickInterval -= readN
				continue
			}
			_, baseLen := h.handler.BaseLen()
			h.clickInterval = h.handler.MessageLength(h.halfBuffer) - baseLen
			h.state = _ScanMsgParse2
		case _ScanMsgParse2:
			readN, readData := utils.ReadFromData(h.clickInterval, data)
			h.halfBuffer = append(h.halfBuffer, readData...)
			if readN == -1 {
				return nil, errors.New("no read data")
			}
			data = data[readN:]
			h.clickInterval -= readN
			if h.clickInterval > 0 {
				continue
			}
			msg := h.allocTor.AllocMessage()
			msg.Reset()
			action, err := h.handler.Unmarshal(h.halfBuffer, msg)
			if err != nil {
				h.allocTor.FreeMessage(msg)
				return nil, err
			}
			err = h.handleAction(action, msg, &allMsg, readData)
			if err != nil {
				h.allocTor.FreeMessage(msg)
				return nil, err
			}
			h.ResetScan()
		}
	}
}

func (h *lRPCTrait) Free(msg *message2.Message) {
	h.allocTor.FreeMessage(msg)
}

func (h *lRPCTrait) Reset() {
	h.ResetScan()
}

func (h *lRPCTrait) ResetScan() {
	h.handler = nil
	h.halfBuffer = h.halfBuffer[:0]
	h.clickInterval = 1
	h.state = _ScanInit
}

func (h *lRPCTrait) deleteNoReadyBuffer(msgId uint64) {
	// 置空/删除Map Key让内存得以回收
	h.noReadyBuffer[msgId] = readyBuffer{}
	delete(h.noReadyBuffer, msgId)
}

// State 下个状态的触发间隔&当前的状态&缓冲区的长度
func (h *lRPCTrait) State() (int, int, int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.clickInterval, h.state, len(h.halfBuffer)
}

func (h *lRPCTrait) handleAction(action Action, msg *message2.Message, allMsg *[]ParserMessage, readData []byte) error {
	switch action {
	case UnmarshalBase:
		buf, ok := h.noReadyBuffer[msg.GetMsgId()]
		if !ok {
			h.noReadyBuffer[msg.GetMsgId()] = readyBuffer{
				MsgId:         msg.GetMsgId(),
				PayloadLength: msg.Length(),
				RawBytes:      h.halfBuffer,
			}
		}
		buf.RawBytes = append(buf.RawBytes, readData...)
		if uint32(len(buf.RawBytes)) == buf.PayloadLength {
			defer h.deleteNoReadyBuffer(msg.GetMsgId())
			msg.Reset()
			err := message2.Unmarshal(buf.RawBytes, msg)
			if err != nil {
				return err
			}
			*allMsg = append(*allMsg, ParserMessage{
				Message: msg,
				Header:  buf.RawBytes[0],
			})
		}
	case UnmarshalComplete:
		*allMsg = append(*allMsg, ParserMessage{
			Message: msg,
			Header:  h.halfBuffer[0],
		})
	}
	return nil
}

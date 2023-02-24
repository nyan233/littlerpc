package msgparser

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/core/container"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/utils"
	"sync"
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
	state int
	// 单次解析数据的起始偏移量
	startOffset int
	// 单次解析数据的结束偏移量
	endOffset     int
	noReadyBuffer map[uint64]readyBuffer
	// 存储半包数据的缓冲区, 只有在读完了一个完整的payload的消息的数据包
	// 才会被直接提升到noReadyBuffer中, noMux类型的数据包则不会被提升到
	// noReadyBuffer中, 将完整的消息读取完毕后直接触发onComplete
	halfBuffer container.ByteSlice
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

func (h *lRPCTrait) ParseOnReader(reader func([]byte) (n int, err error)) (msgs []ParserMessage, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	currentLen := len(h.halfBuffer)
	currentCap := cap(h.halfBuffer)
	h.halfBuffer = h.halfBuffer[:currentCap]
	for i := 0; i < 16; i++ {
		readN, err := reader(h.halfBuffer[currentLen:currentCap])
		if readN > 0 {
			currentLen += readN
		}
		// read full
		if currentLen == currentCap {
			break
		}
		if err != nil {
			break
		}
	}
	h.halfBuffer = h.halfBuffer[:currentLen]
	return h.parseFromHalfBuffer(nil)
}

// Parse io.Reader主要用来标识一个读取到半包的连接, 并不会真正去调用他的方法
func (h *lRPCTrait) Parse(data []byte) (msgs []ParserMessage, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clickInterval == 1 && len(data) == 0 {
		return nil, errors.New("data length == 0")
	}
	return h.parseFromHalfBuffer(data)
}

func (h *lRPCTrait) parseFromHalfBuffer(prepare container.ByteSlice) (msgs []ParserMessage, err error) {
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
	var buf container.ByteSlice
	switch {
	case len(h.halfBuffer) <= 0:
		if prepare.Available() {
			buf = prepare
		}
	case len(h.halfBuffer) > 0:
		buf = h.halfBuffer
		if prepare.Available() {
			buf.Append(prepare)
		}
	}
	for {
		if len(buf) == h.endOffset {
			h.halfBuffer.Reset()
			h.startOffset = 0
			h.endOffset = 0
			return allMsg, nil
		}
		if (len(buf) - h.startOffset) < h.clickInterval {
			h.halfBuffer = h.halfBuffer[:len(buf)-h.startOffset]
			copy(h.halfBuffer, buf[h.startOffset:])
			h.endOffset = h.endOffset - h.startOffset
			h.startOffset = 0
			return allMsg, nil
		}
		switch h.state {
		case _ScanInit:
			err := h.handleScanInit(&buf, &h.startOffset, &h.endOffset, &allMsg)
			if err != nil {
				return nil, err
			}
		case _ScanMsgParse1:
			next, err := h.handleScanParse1(&buf, &h.startOffset, &h.endOffset, &allMsg)
			if err != nil {
				return nil, err
			}
			if !next {
				continue
			}
		case _ScanMsgParse2:
			_, err := h.handleScanParse2(&buf, &h.startOffset, &h.endOffset, &allMsg)
			if err != nil {
				return nil, err
			}
		}
	}
}

func (h *lRPCTrait) handleScanInit(buf *container.ByteSlice, start, end *int, allMsg *[]ParserMessage) (err error) {
	if handler := GetHandler((*buf)[*start]); handler != nil {
		h.handler = handler
	} else {
		return errors.New(fmt.Sprintf("MagicNumber no MessageHandler -> %d", (*buf)[0]))
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
		action, err := h.handler.Unmarshal(*buf, msg)
		if err != nil {
			return err
		}
		*end = len(*buf)
		err = h.handleAction(action, *buf, msg, allMsg, nil)
		if err != nil {
			return err
		}
		return nil
	}
	h.clickInterval = baseLen - 1
	*end++
	return nil
}

func (h *lRPCTrait) handleScanParse1(buf *container.ByteSlice, start, end *int, allMsg *[]ParserMessage) (next bool, err error) {
	readN, _ := utils.ReadFromData(h.clickInterval, (*buf)[*start:])
	*end += readN
	if readN < h.clickInterval {
		h.clickInterval -= readN
		return
	}
	_, baseLen := h.handler.BaseLen()
	h.clickInterval = h.handler.MessageLength((*buf)[*start:*end]) - baseLen
	h.state = _ScanMsgParse2
	next = true
	return
}

func (h *lRPCTrait) handleScanParse2(buf *container.ByteSlice, start, end *int, allMsg *[]ParserMessage) (next bool, err error) {
	readN, readData := utils.ReadFromData(h.clickInterval, (*buf)[*start:])
	if readN == -1 {
		return false, errors.New("no read buf")
	}
	*end += readN
	h.clickInterval -= readN
	if h.clickInterval > 0 {
		return
	}
	msg := h.allocTor.AllocMessage()
	msg.Reset()
	action, err := h.handler.Unmarshal((*buf)[*start:*end], msg)
	if err != nil {
		h.allocTor.FreeMessage(msg)
		return false, err
	}
	err = h.handleAction(action, *buf, msg, allMsg, readData)
	if err != nil {
		h.allocTor.FreeMessage(msg)
		return false, err
	}
	h.ResetScan()
	*start = *end
	return
}

func (h *lRPCTrait) Free(msg *message2.Message) {
	h.allocTor.FreeMessage(msg)
}

func (h *lRPCTrait) Reset() {
	h.ResetScan()
}

func (h *lRPCTrait) ResetScan() {
	h.handler = nil
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

func (h *lRPCTrait) handleAction(action Action, buf container.ByteSlice, msg *message2.Message, allMsg *[]ParserMessage, readData []byte) error {
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
			Header:  buf[h.startOffset],
		})
	}
	return nil
}

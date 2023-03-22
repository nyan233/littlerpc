package msgparser

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/core/container"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
)

type readyBuffer struct {
	MsgId         uint64
	PayloadLength uint32
	RawBytes      []byte
}

func (b *readyBuffer) IsUnmarshal() bool {
	return b.MsgId != 0 && b.PayloadLength != 0
}

// TODO: 取消不必要的锁, 流数据都是顺序到来的, 锁没必要
// lRPCTrait 特征化Parser, 根据Header自动选择适合的Handler
type lRPCTrait struct {
	// 简单的分配器接口, 用于分配可复用的Message
	allocTor Allocator
	// 下一个状态的触发间隔, 也就是距离转移到下一个状态需要读取的数据量
	clickInterval int
	// 当前parser选中的handler
	handler MessageHandler
	// 当前在状态机中处于的状态
	state int
	// 单次解析数据的起始偏移量
	startOffset int
	// 行内指针, 指示在start-end中目前的解析处于哪个位置
	linePtr int
	// 单次解析数据的结束偏移量
	endOffset int
	// 消息的最小解析长度, 包括1Byte的Header
	msgBaseLen int
	// 用于存储未完成的消息, 用于mux
	noReadyBuffer map[uint64]readyBuffer
	// 存储半包数据的缓冲区, 只有在读完了一个完整的payload的消息的数据包
	// 才会被直接提升到noReadyBuffer中, noMux类型的数据包则不会被提升到
	// noReadyBuffer中, 将完整的消息读取完毕后直接触发onComplete
	halfBuffer container.ByteSlice
}

func NewLRPCTrait(allocTor Allocator, bufSize uint32) Parser {
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
	currentLen := h.halfBuffer.Len()
	currentCap := h.halfBuffer.Cap()
	h.halfBuffer = h.halfBuffer[:currentCap]
	readN, err := reader(h.halfBuffer[currentLen:currentCap])
	if err != nil {
		return nil, err
	}
	h.halfBuffer = h.halfBuffer[:currentLen+readN]
	return h.parseFromHalfBuffer()
}

// Parse io.Reader主要用来标识一个读取到半包的连接, 并不会真正去调用他的方法
func (h *lRPCTrait) Parse(data []byte) (msgs []ParserMessage, err error) {
	if h.clickInterval == 1 && len(data) == 0 {
		return nil, errors.New("data length == 0")
	}
	h.halfBuffer.Append(data)
	return h.parseFromHalfBuffer()
}

func (h *lRPCTrait) memSwap() {
	if !(h.startOffset+h.clickInterval > h.halfBuffer.Cap()) {
		return
	}

}

func (h *lRPCTrait) parseFromHalfBuffer() (msgs []ParserMessage, err error) {
	allMsg := h.allocTor.AllocContainer()
	allMsg.Reset()
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
	var ableParse bool
	var parseInterrupt bool
	for {
		if parseInterrupt {
			break
		}
		if ableParse && h.clickInterval > h.halfBuffer.Len()-h.startOffset {
			break
		}
		// scan all
		if h.halfBuffer.Len() == h.endOffset {
			h.halfBuffer.Reset()
			h.startOffset = 0
			h.endOffset = 0
			h.linePtr = 0
			h.msgBaseLen = 0
			return allMsg, nil
		}
		switch h.state {
		case _ScanInit:
			err := h.handleScanInit(&allMsg)
			if err != nil {
				return nil, err
			}
		case _ScanMsgParse1:
			next, err := h.handleScanParse1(&allMsg)
			if err != nil {
				return nil, err
			}
			if !next {
				parseInterrupt = true
			} else {
				ableParse = true
			}
		case _ScanMsgParse2:
			next, err := h.handleScanParse2(&allMsg)
			if err != nil {
				return nil, err
			}
			if !next {
				parseInterrupt = true
			} else {
				ableParse = true
			}
		}
	}
	// 最后的数据不满足长度要求则可以搬迁数据, 至少要经过一次完整的解析
	if ableParse && (h.halfBuffer.Len()-h.startOffset < h.clickInterval) && h.startOffset > 0 {
		oldBuffer := h.halfBuffer
		h.halfBuffer = h.halfBuffer[:h.halfBuffer.Len()-h.startOffset]
		copy(h.halfBuffer, oldBuffer[h.startOffset:])
		h.endOffset = h.endOffset - h.startOffset
		h.startOffset = 0
		h.linePtr = h.endOffset
		h.msgBaseLen = 0
	}
	return allMsg, nil
}

func (h *lRPCTrait) handleScanInit(allMsg *container.Slice[ParserMessage]) (err error) {
	if handler := GetHandler(h.halfBuffer[h.startOffset]); handler != nil {
		h.handler = handler
	} else {
		return errors.New(fmt.Sprintf("MagicNumber no MessageHandler -> %d", (h.halfBuffer)[0]))
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
		action, err := h.handler.Unmarshal(h.halfBuffer, msg)
		if err != nil {
			return err
		}
		h.linePtr = h.halfBuffer.Len()
		h.endOffset = h.halfBuffer.Len()
		err = h.handleAction(action, h.halfBuffer, msg, allMsg, nil)
		if err != nil {
			return err
		}
		return nil
	}
	h.msgBaseLen = baseLen
	h.clickInterval = baseLen - 1
	h.endOffset++
	h.linePtr++
	return nil
}

func (h *lRPCTrait) handleScanParse1(allMsg *container.Slice[ParserMessage]) (next bool, err error) {
	interval := h.halfBuffer.Len() - h.linePtr
	if interval < 0 {
		return false, errors.New("no read buf")
	}
	if interval < h.clickInterval {
		return false, nil
	}
	interval = h.clickInterval
	h.linePtr += interval
	h.endOffset += interval
	h.clickInterval = 0
	_, baseLen := h.handler.BaseLen()
	h.clickInterval = h.handler.MessageLength(h.halfBuffer[h.startOffset:h.endOffset]) - baseLen
	h.state = _ScanMsgParse2
	next = true
	return
}

func (h *lRPCTrait) handleScanParse2(allMsg *container.Slice[ParserMessage]) (next bool, err error) {
	interval := h.halfBuffer.Len() - h.linePtr
	if interval < 0 {
		return false, errors.New("no read buf")
	}
	if interval < h.clickInterval {
		return false, nil
	}
	interval = h.clickInterval
	h.linePtr += interval
	h.endOffset += interval
	h.clickInterval = 0
	msg := h.allocTor.AllocMessage()
	msg.Reset()
	action, err := h.handler.Unmarshal(h.halfBuffer[h.startOffset:h.endOffset], msg)
	if err != nil {
		h.allocTor.FreeMessage(msg)
		return false, err
	}
	err = h.handleAction(action, h.halfBuffer, msg, allMsg, h.halfBuffer[h.startOffset+h.msgBaseLen:h.endOffset])
	if err != nil {
		h.allocTor.FreeMessage(msg)
		return false, err
	}
	h.ResetScan()
	h.startOffset = h.endOffset
	next = true
	return
}

func (h *lRPCTrait) FreeMessage(msg *message2.Message) {
	h.allocTor.FreeMessage(msg)
}

func (h *lRPCTrait) FreeContainer(c []ParserMessage) {
	h.allocTor.FreeContainer(c)
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

func (h *lRPCTrait) handleAction(action Action, buf container.ByteSlice, msg *message2.Message, allMsg *container.Slice[ParserMessage], readData []byte) error {
	switch action {
	case UnmarshalBase:
		readBuf, ok := h.noReadyBuffer[msg.GetMsgId()]
		if !ok {
			readBuf = readyBuffer{
				MsgId:         msg.GetMsgId(),
				PayloadLength: msg.Length(),
				RawBytes:      buf,
			}
		} else {
			readBuf.RawBytes = append(readBuf.RawBytes, readData...)
		}
		if uint32(len(readBuf.RawBytes)) == readBuf.PayloadLength {
			defer h.deleteNoReadyBuffer(msg.GetMsgId())
			msg.Reset()
			err := message2.Unmarshal(readBuf.RawBytes, msg)
			if err != nil {
				return err
			}
			*allMsg = append(*allMsg, ParserMessage{
				Message: msg,
				Header:  readBuf.RawBytes[0],
			})
		} else if !ok {
			readBuf.RawBytes = append([]byte{}, readData...)
			// mux中的消息不能一次性序列化完成则释放预分配的msg
			h.allocTor.FreeMessage(msg)
		}
		h.noReadyBuffer[msg.GetMsgId()] = readBuf
	case UnmarshalComplete:
		*allMsg = append(*allMsg, ParserMessage{
			Message: msg,
			Header:  buf[h.startOffset],
		})
	}
	return nil
}

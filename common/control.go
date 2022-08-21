package common

import (
	"errors"
	"github.com/nyan233/littlerpc/container"
	"github.com/nyan233/littlerpc/protocol"
	"io"
	"sync"
)

type WriteLocker interface {
	sync.Locker
	io.Writer
}

type ReadLocker interface {
	sync.Locker
	io.Reader
}

func CheckConnWrite(conn io.Writer, p []byte) error {
	wn, err := conn.Write(p)
	if err != nil {
		return err
	}
	if wn != len(p) {
		return errors.New("write bytes not equal")
	}
	return nil
}

func CheckConnRead(n int, err error) error {
	if err != nil {
		return err
	}
	return nil
}

// ReadControl data中的数据应该被置零
func ReadControl(c io.Reader, data []byte) error {
	var readCount int
	for {
		readN, err := c.Read(data[readCount:])
		if err != nil {
			return err
		}
		readCount += readN
		if readCount >= len(data) {
			break
		}
	}
	return nil
}

func WriteControl(c io.Writer, data []byte) error {
	var writeCount int
	for {
		writeN, err := c.Write(data[writeCount:])
		if err != nil {
			return err
		}
		writeCount += writeN
		if writeCount >= len(data) {
			break
		}
	}
	return nil
}

// MuxWriteAll muxMsg是预定义的Mux载荷
// mmBytes 是供拷贝数据的缓冲区
// mBytes是要写入的数据
// startFn是每次循环开始时都会调用的回调函数,它允许你在开始前做一些检查
func MuxWriteAll(c WriteLocker, muxMsg *protocol.MuxBlock, mmBytes *container.Slice[byte],
	mBytes []byte, startFn func()) error {
	for len(mBytes) > 0 {
		if startFn != nil {
			startFn()
		}
		var sendN int
		if len(mBytes) < protocol.MaxPayloadSizeOnMux {
			sendN = len(mBytes)
		} else {
			sendN = protocol.MaxPayloadSizeOnMux
		}
		mmBytes.Reset()
		muxMsg.Payloads = mBytes[:sendN]
		muxMsg.PayloadLength = uint16(sendN)
		err := protocol.MarshalMuxBlock(muxMsg, mmBytes)
		if err != nil {
			return err
		}
		c.Lock()
		err = WriteControl(c, *mmBytes)
		if err != nil {
			c.Unlock()
			return err
		}
		c.Unlock()
		mBytes = mBytes[sendN:]
	}
	return nil
}

// MuxReadAll checkPoint在Lock之前被执行,如果需要goroutine安全则需要在匿名函数内编写Lock/Unlock
// oneComplete在完成一次MuxMessage接收时被调用
// 在Client模式下mmBytes作为读的缓冲,其容量必须 >= protocol.MuxMessageBlockSize
// Server模式并没有检查点, 因为检查点对于Server来说并没有意义
// oneComplete回调中被传入的p不应该将其缓存, 如果后续会使用到则应该拷贝它, 因为p大部分的情况下只是mmBytes一部分数据的拷贝
//
//	NOTE checkPoint不是必选项,oneComplete回调必须被注册,否则会调用panic
//	NOTE 不可能出现载荷长度为0的情况,LittleRpc的协议规定了Message的最小长度
func MuxReadAll(c ReadLocker, mmBytes container.Slice[byte],
	checkPoint func(c ReadLocker) bool, oneComplete func(mm protocol.MuxBlock) bool) error {

	if oneComplete == nil {
		panic("no pass by oneComplete callback")
	}
	if mmBytes.Cap() < protocol.MuxMessageBlockSize {
		panic("buffer capacity less than protocol.MuxMessageBlockSize")
	}
	return muxReadAll(c, mmBytes, checkPoint, oneComplete)
}

func muxReadAll(c ReadLocker, mmBytes container.Slice[byte],
	checkPoint func(c ReadLocker) bool, oneComplete func(mm protocol.MuxBlock) bool) error {

	c.Lock()
	bytes := mmBytes
	for {
		if checkPoint != nil && !checkPoint(c) {
			c.Unlock()
			break
		}
		var muxMsg protocol.MuxBlock
		if bytes.Len() < protocol.MuxBlockBaseLen {
			// Server OnMessage响应的数据包不满足基本长度
			bytes = bytes[bytes.Len():protocol.MuxBlockBaseLen]
			err := ReadControl(c, bytes)
			if err != nil {
				c.Unlock()
				return err
			}
		}
		err := protocol.UnmarshalMuxBlock(bytes, &muxMsg)
		if err != nil {
			c.Unlock()
			return err
		}
		if muxMsg.PayloadLength == 0 {
			c.Unlock()
			return errors.New("message length is zero")
		}
		// 未读出一个完整载荷
		if muxMsg.Payloads == nil || int(muxMsg.PayloadLength) > muxMsg.Payloads.Len() {
			if muxMsg.PayloadLength > protocol.MaxPayloadSizeOnMux {
				bytes = (bytes)[:protocol.MaxPayloadSizeOnMux]
			} else {
				bytes = (bytes)[:muxMsg.PayloadLength]
			}
			err := ReadControl(c, bytes)
			if err != nil {
				c.Unlock()
				return err
			}
			muxMsg.Payloads = bytes
			oneComplete(muxMsg)
			break
		} else if muxMsg.Payloads != nil && int(muxMsg.PayloadLength) < muxMsg.Payloads.Len() {
			// 数据中包含下一个载荷,下一个载荷可能是完整包,也可能是半包
			// 尽快调用oneComplete()因为拷贝数据之后会导致原数据被修改
			muxMsg.Payloads = muxMsg.Payloads[:muxMsg.PayloadLength]
			oneComplete(muxMsg)
			nextBytes := muxMsg.Payloads[muxMsg.PayloadLength:]
			copy(bytes, nextBytes)
		} else {
			// 仅仅包含一个完整的载荷
			oneComplete(muxMsg)
			break
		}
	}
	c.Unlock()
	return nil
}

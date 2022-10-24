package control

import (
	"errors"
	"fmt"
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/protocol/mux"
	"io"
	"syscall"
)

const MessageTrace = 0

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

// ReadControl
//
//	NOTE: data中的数据应该被置零
//	NOTE: Read应该检测属于Nio的中断错误, 因为不确定其它框架中是否实现
//	NOTE: 这些错误的过滤
func ReadControl(c io.Reader, data []byte) error {
	var readCount int
	for {
		readN, err := c.Read(data[readCount:])
		if err != nil && !(err == syscall.EAGAIN || err == syscall.EINTR || err == syscall.EWOULDBLOCK) {
			return err
		} else if readN < 0 {
			readN = 0
		}
		readCount += readN
		if readCount == len(data) {
			break
		}
	}
	return nil
}

// WriteControl
//
//	NOTE: Write应该检测属于Nio的中断错误, 因为不确定其它框架中是否实现
//	NOTE: 这些错误的过滤
func WriteControl(c io.Writer, data []byte) error {
	var writeCount int
	for {
		writeN, err := c.Write(data[writeCount:])
		if err != nil && !(err == syscall.EAGAIN || err == syscall.EINTR || err == syscall.EWOULDBLOCK) {
			return err
		} else if writeN < 0 {
			writeN = 0
		}
		writeCount += writeN
		if writeCount == len(data) {
			break
		}
	}
	return nil
}

// MuxWriteAll muxMsg是预定义的Mux载荷
// mmBytes 是供拷贝数据的缓冲区
// mBytes是要写入的数据
// startFn是每次循环开始时都会调用的回调函数,它允许你在开始前做一些检查
func MuxWriteAll(c io.Writer, muxMsg *mux.Block, mmBytes *container.Slice[byte],
	mBytes []byte, startFn func()) error {
	if mmBytes == nil {
		var tmp container.Slice[byte]
		if len(mBytes) <= mux.MaxBlockSize {
			tmp = make([]byte, 0, len(mBytes))
		} else {
			tmp = make([]byte, mux.MaxBlockSize)
		}
		mmBytes = &tmp
	}
	for len(mBytes) > 0 {
		if startFn != nil {
			startFn()
		}
		var sendN int
		if len(mBytes) < mux.MaxPayloadSizeOnMux {
			sendN = len(mBytes)
		} else {
			sendN = mux.MaxPayloadSizeOnMux
		}
		mmBytes.Reset()
		muxMsg.Payloads = mBytes[:sendN]
		muxMsg.PayloadLength = uint16(sendN)
		err := mux.Marshal(muxMsg, mmBytes)
		if err != nil {
			return err
		}
		if MessageTrace > 0 {
			fmt.Println("Write ", *muxMsg)
			fmt.Println("Write Raw", string(*mmBytes))
		}
		err = WriteControl(c, *mmBytes)
		if err != nil {
			return err
		}
		mBytes = mBytes[sendN:]
	}
	return nil
}

package control

import (
	"io"
	"syscall"
)

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

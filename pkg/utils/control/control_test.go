package control

import (
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/stretchr/testify/assert"
	"syscall"
	"testing"
)

type testInter struct {
	clickInterval  int
	oneCompleteNum int
	throw          bool
	callCount      int
}

func (t *testInter) Read(p []byte) (n int, err error) {
	defer func() {
		t.callCount++
	}()
	var readN int
	if len(p) <= t.oneCompleteNum {
		readN = len(p)
	} else {
		readN = t.oneCompleteNum
	}
	if t.callCount == t.clickInterval {
		t.callCount = 0
		if t.throw {
			return readN, syscall.EINPROGRESS
		}
		switch random.FastRandN(3) {
		case 0:
			return readN, syscall.EAGAIN
		case 1:
			return readN, syscall.EINTR
		case 2:
			return readN, syscall.EWOULDBLOCK
		}
	}
	return readN, nil
}

func (t *testInter) Write(p []byte) (n int, err error) {
	defer func() {
		t.callCount++
	}()
	var writeN int
	if len(p) <= t.oneCompleteNum {
		writeN = len(p)
	} else {
		writeN = t.oneCompleteNum
	}
	if t.callCount == t.clickInterval {
		t.callCount = 0
		if t.throw {
			return writeN, syscall.EINPROGRESS
		}
		switch random.FastRandN(3) {
		case 0:
			return writeN, syscall.EAGAIN
		case 1:
			return writeN, syscall.EINTR
		case 2:
			return writeN, syscall.EWOULDBLOCK
		}
	}
	return writeN, nil
}

func TestWriteControl(t *testing.T) {
	const (
		ClickInterval  = 10
		OneCompleteNum = 100
	)
	bytes := make([]byte, 1024*1024)
	inter := &testInter{
		clickInterval:  ClickInterval,
		oneCompleteNum: OneCompleteNum,
	}
	assert.Equal(t, ReadControl(inter, bytes), nil)
	inter.callCount = 0
	assert.Equal(t, WriteControl(inter, bytes), nil)
	inter.callCount = 0
	inter.throw = true
	assert.NotEqual(t, ReadControl(inter, bytes), nil)
	assert.NotEqual(t, WriteControl(inter, bytes), nil)
}

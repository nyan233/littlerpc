package client

import (
	"github.com/nyan233/littlerpc/protocol/mux"
	"sync/atomic"
)

func getConnFromMux(c *Client) *lockConn {
	count := atomic.LoadInt64(&c.concurrentConnCount)
	atomic.AddInt64(&c.concurrentConnCount, 1)
	return c.concurrentConnect[count%int64(len(c.concurrentConnect))]
}

func getSendBlockBytes(sendBlockCount int, p []byte) []byte {
	start := (sendBlockCount - 1) * mux.MuxMessageBlockSize
	end := sendBlockCount * mux.MuxMessageBlockSize
	if end > len(p) {
		end = len(p)
	}
	return p[start:end]
}

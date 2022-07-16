package client

import "sync/atomic"

func getConnFromMux(c *Client) lockConn {
	count := atomic.LoadInt64(&c.concurrentConnCount)
	atomic.AddInt64(&c.concurrentConnCount, 1)
	return c.concurrentConnect[count%int64(len(c.concurrentConnect))]
}

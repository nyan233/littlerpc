package server

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	ev := new(event)
	testEvent(t, int(_Start), ev)
	testEvent(t, int(_Stop), ev)
	testEvent(t, int(_Restart), ev)
}

func testEvent(t *testing.T, ev int, i *event) {
	const (
		WaitSize = 200
	)
	done := make(chan struct{})
	go func() {
		if !i.Entry(ev) {
			t.Error("entry event failed")
		}
		done <- struct{}{}
		select {
		case <-done:
			if !i.Complete(ev) {
				t.Error("complete event failed")
			}
			done <- struct{}{}
		}
	}()
	<-done
	var wg sync.WaitGroup
	wg.Add(WaitSize)
	for j := 0; j < WaitSize; j++ {
		go func() {
			wg.Done()
			w, ack, ok := i.Wait()
			assert.Equal(t, ok, true)
			assert.Equal(t, <-w, ev)
			ack()
			wg.Done()
		}()
	}
	wg.Wait()
	wg.Add(WaitSize)
	time.AfterFunc(time.Second*2, func() {
		done <- struct{}{}
	})
	wg.Wait()
	<-done
}

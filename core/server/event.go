package server

import "sync"

type eventList struct {
	done chan int
	next *eventList
}

type event struct {
	mu sync.Mutex
	ev int
	// 用于确认所有的等待已经完成
	ackWait sync.WaitGroup
	link    *eventList
}

func newEvent() *event {
	return &event{
		ev: -1,
	}
}

// Entry 进入一个事件
func (e *event) Entry(ev int) bool {
	if !e.mu.TryLock() {
		return false
	}
	defer e.mu.Unlock()
	if e.ev > 0 {
		return false
	}
	e.ev = ev
	return true
}

// Complete 完成一个事件, 返回时所有等待者已经完成了任务
func (e *event) Complete(ev int) bool {
	if !e.mu.TryLock() {
		return false
	}
	defer e.mu.Unlock()
	if e.ev < 0 || e.ev != ev {
		return false
	}
	for e.link != nil {
		e.link.done <- ev
		e.link = e.link.next
	}
	e.ev = -1
	e.ackWait.Wait()
	return true
}

// Wait 等待一个事件完成, done用于接收(*event).Complete()是否已经被调用
// ack用于确认一个等待者是否已经完成了操作, 未进入事件时不允许等待
func (e *event) Wait() (done chan int, ack func(), ok bool) {
	e.mu.Lock()
	if e.ev < 0 {
		e.mu.Unlock()
		return nil, nil, false
	}
	defer e.mu.Unlock()
	e.link = &eventList{
		done: make(chan int, 1),
		next: e.link,
	}
	done = e.link.done
	e.ackWait.Add(1)
	var called bool
	ack = func() {
		if called {
			panic("ackWait already called")
		}
		called = true
		e.ackWait.Done()
	}
	ok = true
	return
}

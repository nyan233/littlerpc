package container

import (
	"sync"
	"sync/atomic"
)

type RCUMapElement[Key comparable, Val any] struct {
	Key   Key
	Value Val
}

type RCUDeleteNode[Val any] struct {
	Value Val
	Ok    bool
}

// RCUMap 这个Map的实现只适合少量key-value, 或者几乎无写的场景
// 在大量key-value时拷贝数据的开销很大
type RCUMap[Key comparable, Val any] struct {
	mu      sync.Mutex // 串行写入操作, 读取操作不需要上锁
	pointer atomic.Pointer[map[Key]Val]
}

func NewRCUMap[K comparable, V any]() *RCUMap[K, V] {
	m := new(RCUMap[K, V])
	tmp := make(map[K]V, 128)
	m.pointer.Store(&tmp)
	return m
}

func (R *RCUMap[Key, Val]) LoadOk(key Key) (Val, bool) {
	snapshot := R.pointer.Load()
	val, ok := (*snapshot)[key]
	return val, ok
}

func (R *RCUMap[Key, Val]) Range(fn func(key Key, val Val) bool) {
	snapshot := R.pointer.Load()
	for k, v := range *snapshot {
		if !fn(k, v) {
			break
		}
	}
}

func (R *RCUMap[Key, Val]) Store(key Key, val Val) {
	R.StoreMulti([]RCUMapElement[Key, Val]{{Key: key, Value: val}})
}

func (R *RCUMap[Key, Val]) StoreMulti(kvs []RCUMapElement[Key, Val]) {
	if kvs == nil || len(kvs) == 0 {
		return
	}
	R.mu.Lock()
	defer R.mu.Unlock()
	copyMap := R.copy()
	for _, kv := range kvs {
		copyMap[kv.Key] = kv.Value
	}
	R.pointer.Store(&copyMap)
}

func (R *RCUMap[Key, Val]) DeleteOk(key Key) (Val, bool) {
	val := R.DeleteMulti([]Key{key})
	if val == nil || len(val) == 0 {
		return *new(Val), false
	}
	return val[0].Value, val[0].Ok
}

func (R *RCUMap[Key, Val]) Delete(key Key) {
	R.DeleteOk(key)
}

func (R *RCUMap[Key, Val]) DeleteMulti(keys []Key) []RCUDeleteNode[Val] {
	if keys == nil || len(keys) == 0 {
		return nil
	}
	R.mu.Lock()
	defer R.mu.Unlock()
	copyMap := R.copy()
	values := make([]RCUDeleteNode[Val], 0, len(keys))
	for _, key := range keys {
		val, ok := copyMap[key]
		values = append(values, RCUDeleteNode[Val]{
			Value: val,
			Ok:    ok,
		})
		if !ok {
			continue
		}
		delete(copyMap, key)
	}
	R.pointer.Store(&copyMap)
	return values
}

func (R *RCUMap[Key, Val]) Len() int {
	return len(*R.pointer.Load())
}

func (R *RCUMap[Key, Val]) copy() map[Key]Val {
	snapshot := *R.pointer.Load()
	copyMap := make(map[Key]Val, len(snapshot))
	for k, v := range snapshot {
		copyMap[k] = v
	}
	return copyMap
}

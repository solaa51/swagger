package kvLibrary

import "sync/atomic"

//kv并发读写库

type KVMap[K comparable, V any] struct {
	state atomic.Uint32 //0待处理 1写入/删除 2读取
	count atomic.Int64  //数量
	kv    map[K]V
}

func NewKVMap[K comparable, V any](cap int) *KVMap[K, V] {
	return &KVMap[K, V]{
		kv: make(map[K]V, cap),
	}
}

func (m *KVMap[K, V]) Set(k K, v V) {
	for {
		if m.state.CompareAndSwap(0, 1) {
			m.kv[k] = v
			m.count.Add(1)
			m.state.CompareAndSwap(1, 0)
			return
		}
	}
}

func (m *KVMap[K, V]) Get(k K) (V, bool) {
	for {
		if m.state.CompareAndSwap(0, 2) {
			v, ok := m.kv[k]
			m.state.CompareAndSwap(2, 0)
			return v, ok
		}
	}
}

func (m *KVMap[K, V]) Del(k K) {
	for {
		if m.state.CompareAndSwap(0, 1) {
			delete(m.kv, k)
			m.count.Add(1)
			m.state.CompareAndSwap(1, 0)
			return
		}
	}
}

func (m *KVMap[K, V]) Len() int {
	return int(m.count.Load())
}

func (m *KVMap[K, V]) Clear() {
	for {
		if m.state.CompareAndSwap(0, 1) {
			clear(m.kv)
			//m.kv = make(map[K]V)
			m.count.Store(0)
			m.state.CompareAndSwap(1, 0)
			return
		}
	}
}

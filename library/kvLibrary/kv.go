package kvLibrary

import (
	"fmt"
	"maps"
	"sync/atomic"
)

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

// Items 获取所有数据 复制后返回
func (m *KVMap[K, V]) Items() map[K]V {
	for {
		if m.state.CompareAndSwap(0, 2) {
			p := make(map[K]V, m.count.Load())
			maps.Copy(p, m.kv)
			m.state.CompareAndSwap(2, 0)
			return p
		}
	}
}

// SetOnce 仅允许单次写入
func (m *KVMap[K, V]) SetOnce(k K, v V) bool {
	b := false
	for {
		if m.state.CompareAndSwap(0, 1) {
			if _, ok := m.kv[k]; !ok {
				m.kv[k] = v
				m.count.Add(1)
				b = true
				fmt.Println("kv库写入成功", k, v)
			}
			m.state.CompareAndSwap(1, 0)
			break
		}
	}

	return b
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

package kvLibrary

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

/*
*
//go test -bench . -benchtime=5s

go test -bench . -benchmem

	goos: darwin
	goarch: amd64
	pkg: github.com/solaa51/swagger/library/kvLibrary
	cpu: Intel(R) Core(TM) i7-8700B CPU @ 3.20GHz
	BenchmarkGoSyncMapReadsOnly-12             58581             21620 ns/op
	BenchmarkAtomicMapReadsOnly-12               627           1870974 ns/op
	BenchmarkAtomic-12                      59450377                18.22 ns/op
	BenchmarkMap-12                         60533264                19.90 ns/op
	BenchmarkSyncMap-12                     13542068                86.99 ns/op
	PASS
	ok      github.com/solaa51/swagger/library/kvLibrary    6.443s
*/
var epochs uintptr = 1 << 12

// 21620 ns/op 以这个为基准 需要超越这个 才有意义
func BenchmarkGoSyncMapReadsOnly(b *testing.B) {
	b.ResetTimer()
	m := &sync.Map{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < epochs; i++ {
				m.Store(i, i)
			}

			for i := uintptr(0); i < epochs; i++ {
				j, _ := m.Load(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

type strInt int

func (s strInt) String() string {
	return strconv.Itoa(int(s))
}

func BenchmarkAtomicShardMapReadsOnly(b *testing.B) {
	b.ResetTimer()
	m := NewShardKV[strInt, int](24, int(epochs))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < epochs; i++ {
				m.Set(strInt(int(i)), int(i))
			}

			for i := uintptr(0); i < epochs; i++ {
				j, _ := m.Get(strInt(int(i)))
				if j != int(i) {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkAtomicMapReadsOnly(b *testing.B) {
	b.ResetTimer()
	m := NewKVMap[int, int](int(epochs))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < epochs; i++ {
				m.Set(int(i), int(i))
			}

			for i := uintptr(0); i < epochs; i++ {
				j, _ := m.Get(int(i))
				if j != int(i) {
					b.Fail()
				}
			}
		}
	})
}

type mm struct {
	v  atomic.Uint32 //0待处理 1写入 2读取
	kv map[int]int
}

func (m *mm) set(k int, v int) {
	for {
		if m.v.CompareAndSwap(0, 1) {
			m.kv[k] = v
			m.v.CompareAndSwap(1, 0)
			return
		}
	}
}

func (m *mm) Get(k int) int {
	for {
		if m.v.CompareAndSwap(0, 2) {
			v := m.kv[k]
			m.v.CompareAndSwap(2, 0)
			return v
		}
	}
}

// 19.16 ns/op
func BenchmarkAtomic(b *testing.B) {
	b.ResetTimer()
	m := mm{
		kv: make(map[int]int),
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < epochs; i++ {
				m.set(int(i), int(i))
			}

			for i := uintptr(0); i < epochs; i++ {
				j := m.Get(int(i))
				if j != int(i) {
					b.Fail()
				}
			}
		}
	})
}

type px struct {
	mu sync.Mutex
	v  map[int]int
}

func (p *px) set(k int, v int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.v[k] = v
}

func (p *px) get(k int) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.v[k]
}

// 普通map 19.80 ns/op
func BenchmarkMutexMap(b *testing.B) {
	b.ResetTimer()
	p := px{
		v: make(map[int]int),
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < epochs; i++ {
				p.set(int(i), int(i))
			}

			for i := uintptr(0); i < epochs; i++ {
				j := p.get(int(i))
				if j != int(i) {
					b.Fail()
				}
			}
		}
	})
}

// syncMap 86.09 ns/op
func BenchmarkSyncMap(b *testing.B) {
	b.ResetTimer()
	p := sync.Map{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < epochs; i++ {
				p.Store(int(i), i)
			}

			for i := uintptr(0); i < epochs; i++ {
				j, _ := p.Load(int(i))
				if j != int(i) {
					b.Fail()
				}
			}
		}
	})
}

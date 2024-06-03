package kvLibrary

import (
	"sync"
	"sync/atomic"
	"testing"
)

/**
go test -bench .
	goos: darwin
	goarch: amd64
	pkg: m
	cpu: Intel(R) Core(TM) i7-8700B CPU @ 3.20GHz
	BenchmarkAtomic-12      62071592                19.16 ns/op
	BenchmarkMap-12         59370681                19.85 ns/op
	BenchmarkSyncMap-12     13730120                86.09 ns/op
	PASS
	ok      m       3.690s
*/

type mm struct {
	v  atomic.Uint32 //0待处理 1写入 2读取
	kv map[string]int
}

func (m *mm) set(k string, v int) {
	for {
		if m.v.CompareAndSwap(0, 1) {
			m.kv[k] = v
			m.v.CompareAndSwap(1, 0)
			return
		}
	}
}

func (m *mm) Get(k string) int {
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
	m := mm{
		kv: make(map[string]int),
	}

	for i := 0; i < b.N; i++ {
		m.set("haah", i)
	}
}

type px struct {
	mu sync.Mutex
	v  map[string]int
}

func (p *px) set(k string, v int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.v[k] = v
}

func (p *px) get(k string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.v[k]
}

// 19.80 ns/op
func BenchmarkMap(b *testing.B) {
	p := px{
		v: make(map[string]int),
	}

	for i := 0; i < b.N; i++ {
		p.set("haah", i)
	}
}

// 86.09 ns/op
func BenchmarkSyncMap(b *testing.B) {
	p := sync.Map{}

	for i := 0; i < b.N; i++ {
		p.Store("haah", i)
	}
}

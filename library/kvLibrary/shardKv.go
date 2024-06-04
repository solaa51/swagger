package kvLibrary

import (
	"fmt"
	"github.com/solaa51/swagger/cFunc"
	"maps"
	"runtime"
)

// 分片kv 需要用到计算fnv32 key必须能转为string才能计算
// 当数据量比较大时 再使用该分片的方式存储
// key 需要实现 String() 方法
// 分片数 默认按照当前cpu线程数*2
// 虽然没有使用到mutex锁，但是最终执行还是要落到cpu的锁上，只是锁的粒度比较小

type ShardKey interface {
	String() string
	comparable
}

type ShardKV[K ShardKey, V any] struct {
	shard    []*KVMap[K, V]
	shardNum int
}

// NewShardKV 创建分片kv
// shardNum 分片数
// perCap 每个分片容量
func NewShardKV[K ShardKey, V any](shardNum int, perCap int) *ShardKV[K, V] {
	if shardNum == 0 {
		shardNum = runtime.NumCPU() * 2
	}

	shard := make([]*KVMap[K, V], shardNum)
	for i := 0; i < shardNum; i++ {
		shard[i] = NewKVMap[K, V](perCap)
	}
	return &ShardKV[K, V]{
		shard:    shard,
		shardNum: shardNum,
	}
}

func (s *ShardKV[K, V]) Set(key K, value V) {
	s.shard[s.calShard(key)].Set(key, value)
}

func (s *ShardKV[K, V]) Get(key K) (V, bool) {
	return s.shard[s.calShard(key)].Get(key)
}

func (s *ShardKV[K, V]) Del(key K) {
	s.shard[s.calShard(key)].Del(key)
}

func (s *ShardKV[K, V]) Clear() {
	for i := 0; i < s.shardNum; i++ {
		s.shard[i].Clear()
	}
}

func (s *ShardKV[K, V]) Len() int {
	var l int
	for i := 0; i < s.shardNum; i++ {
		l += s.shard[i].Len()
	}

	return l
}

func (s *ShardKV[K, V]) Items() map[K]V {
	m := make(map[K]V)
	for i := 0; i < s.shardNum; i++ {
		maps.Copy(m, s.shard[i].Items())

		fmt.Println(i, s.shard[i].Len())
	}

	return m
}

// 分局key计算分片数
func (s *ShardKV[K, V]) calShard(key K) int {
	return cFunc.HashShard(key.String(), s.shardNum)
}

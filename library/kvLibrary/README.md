key value map高并发读写库

#### 当前性能不如官方syncMap 推荐使用官方库

    咱这库 现在性能虽然不咋的 但是可做些扩展 实现部分redis的功能
    后面再把map替换为链表，性能上提升下
    使用循环链表重写 : 循环链表适用于需要频繁插入、删除和修改数据集合的场景
        1. key必须为 comparable 
        2. 得有自动扩容 不够时 自动扩充10个元数据
        3. 得有自动回收空间 超过2/3无效时触发
        4. 外层记录 data总数[包含删除] 有效data数[不包含删除] 当前容量
        5. 数据得有标记删除
        6. 利用atomic.Pointer来操作数据指针
        7. 先搞一个循环链表试试水
        8. 然后尝试atomic操作
        9. 最后整合进当前库

    ^uintptr(0) == -1

## 各种并发kv库使用场景比较 - 线程安全
    1. 官方syncMap 多读少写 [利用map可快速定位]
    2. 

#### 目标使用场景：
    1. 需要频繁读写
    2. 分片数为cpu核数的倍数
    3. 数据量较大

### 使用示例

## 不分片 少量数据使用
```go
package main
func main() {
	iKv := kvLibrary.NewKVMap[int, int](0)
	iKv.Set(1, 1)
	fmt.Println(iKv.Get(1))
}
```

## 分片调用
```go
package main

//自定义string类型 实现String方法
type tStr string
func (t tStr) String() string {
	return string(t)
}

func main() {
	kv := kvLibrary.NewShardKV[tStr, int](0, 0)
	kv.Set("1", 1)

	fmt.Println(kv.Get("1"))
}
```

```go
//并发测试
package main
func main() {
    iKv := kvLibrary.NewShardKV[tStr, int](1, 0)

	testK := tStr("test")
	allNum := 10000

	var wg sync.WaitGroup
	wg.Add(allNum)
	for i := range allNum {
		go func(i int) {
			if i%3 == 0 {
				iKv.Set(testK, i)
			}

			if i%3 == 1 {
				iKv.Get(testK)
			}

			if i%3 == 2 {
				iKv.Del(testK)
			}

			wg.Done()
		}(i)
	}
	wg.Wait()
}

// key需要实现String()方法
type tStr string

func (t tStr) String() string {
	return string(t)
}
```

### https://github.com/alphadose/haxmap 这个库将性能做到了极致

## 只读模式下的性能比较

    goos: darwin
    goarch: amd64
    pkg: github.com/solaa51/swagger/library/kvLibrary
    cpu: Intel(R) Core(TM) i7-8700B CPU @ 3.20GHz
    BenchmarkGoSyncMapReadsOnly-12             55284             22057 ns/op               0 B/op          0 allocs/op
    BenchmarkAtomicShardMapReadsOnly-12        13494             90002 ns/op           15265 B/op       3996 allocs/op
    BenchmarkAtomicMapReadsOnly-12               640           1941287 ns/op               1 B/op          0 allocs/op
    BenchmarkAtomic-12                           577           2023132 ns/op               4 B/op          0 allocs/op
    BenchmarkMutexMap-12                        4221            277431 ns/op               1 B/op          0 allocs/op
    --- FAIL: BenchmarkSyncMap
    FAIL
    exit status 1
    FAIL    github.com/solaa51/swagger/library/kvLibrary    7.587s


单纯读取数据的性能比较：
    go原生syncMap 性能：20885 ns/op               0 B/op          0 allocs/op
    自己的分片KV [cpu核数*2] 性能：90002 ns/op           15265 B/op       3996 allocs/op
        分片数推荐为cpu核数的2数递增 倍数微增 读取性能提升明显，但不会超过 syncMap
    自己不分配KV 1941287 ns/op               1 B/op          0 allocs/op
    原生map+sync.Mutex 性能最差：277431 ns/op               1 B/op          0 allocs/op


## 1:1读写下的性能差异
    goos: darwin
    goarch: amd64
    pkg: github.com/solaa51/swagger/library/kvLibrary
    cpu: Intel(R) Core(TM) i7-8700B CPU @ 3.20GHz
    BenchmarkGoSyncMapReadsOnly-12              7389            163160 ns/op          127060 B/op      11777 allocs/op
    BenchmarkAtomicShardMapReadsOnly-12         5778            228887 ns/op           31210 B/op       7992 allocs/op
    BenchmarkAtomicMapReadsOnly-12               229           5062992 ns/op             722 B/op          0 allocs/op
    BenchmarkAtomic-12                           264           4747591 ns/op            1309 B/op          0 allocs/op
    BenchmarkMutexMap-12                        1852            664912 ns/op             189 B/op          0 allocs/op
    --- FAIL: BenchmarkSyncMap
    FAIL
    exit status 1
    FAIL    github.com/solaa51/swagger/library/kvLibrary    8.115s

## 貌似map本身的性能 已经限制了读写性能 
    不管我这个怎么弄，本质还是map 

## 若要继续提升，就得用链表 + 原子操作
    使用链表 查找元素就是个关键问题了

    int类型的hash值就是他自己的 无符号类型返回

    如果不将分片对外，采用内部计数，这样就可以不需要hash计算
        或者保留分片可允许外部扩展
        内部采用计数法写入数据
        干掉hash key只需限定为 基础类型和指针就可以了
        我草 查找时 没有hash又搞不定哪个分组了

    不能再使用map 将key和value都放到struct下。利用空间换时间

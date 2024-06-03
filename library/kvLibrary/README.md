key value map高并发读写库

### 使用示例

## 不分片 少量数据使用
```go
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

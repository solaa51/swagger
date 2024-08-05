// SingleFlight 防止缓存击穿
//
// 适用场景: key的生成需要谨慎
//
// 1.单机模式下的生成缓存；如果分布式模式下，请使用其他方式
//
// 2.将多个请求合并成一个请求,比如post短时多次提交；get同一内容

// 从cFunc中独立出来，方便分组，相同key在不同&SingleFlight{}实例下，可区分

```go
package main

import (
	"context"
	"fmt"
	"github.com/solaa51/swagger/library/singleFlight"
	"time"
)

func main() {
	cG := singleFlight.SingleFlight{}

	ctx := context.Background()
	ctx1 := context.WithValue(ctx, "test", "test1")
	ctx2 := context.Background()
	ctx2 = context.WithValue(ctx2, "test", "test2")

	go func() {
		ret, err := singleFlight.SingleRun(ctx1, "test", jisuan)
		fmt.Println("go func调用", ret.(int64), err)
	}()

	go func() {
		cgRet, err := cG.SingleRun(ctx2, "test", jisuan)
		fmt.Println("go func自定义调用", cgRet.(int64), err)
	}()

	go func() {
		ret, err := singleFlight.SingleRun(ctx1, "test", jisuan)
		fmt.Println("go func调用", ret.(int64), err)
	}()

	go func() {
		cgRet, err := cG.SingleRun(ctx2, "test", jisuan)
		fmt.Println("go func自定义调用", cgRet.(int64), err)
	}()

	go func() {
		ret, err := singleFlight.SingleRun(ctx1, "test", jisuan)
		fmt.Println("go func调用", ret.(int64), err)
	}()

	go func() {
		cgRet, err := cG.SingleRun(ctx2, "test", jisuan)
		fmt.Println("go func自定义调用", cgRet.(int64), err)
	}()

	time.Sleep(time.Second * 10)
}

func jisuan() (any, error) {
	time.Sleep(time.Second * 2)
	return time.Now().UnixMicro(), nil
}
```


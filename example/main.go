package main

import (
	"github.com/solaa51/swagger/appServer"
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/example/controller"
	"github.com/solaa51/swagger/handle/handleFuncParse"
	router "github.com/solaa51/swagger/routerV2"
)

func init() {
	//配置路由规则
	r := &router.RouteParse{}
	r.BindFunc("welcome/index2", func(ctx *context.Context) {
		ctx.RetData = "hello world2"
	})

	(&router.RouteParse{}).Prefix("gameApi").BindFunc("gameMall/index", func(ctx *context.Context) {
		ctx.RetData = "hello world2"
	})

	(&router.RouteParse{}).BindStructs(
		func() handleFuncParse.Control { return &controller.Auth{} },
	)
}

func main() {
	appServer.Run()
	// 中间件 或 路由 是否有必要支持  局部方法 限流处理
	// 路由匹配能否 用 具体的值 减少指针的使用  可以减少gc的负担
	// 无解的情况下 将对象拆分为值和地址两部分 减少垃圾收集时间
	// 减少程序中指针的数量会减少堆分配的数量

	// 试试将segment改为没有指针的结构 这样将无GC扫描 无解啊  go限定无法获取map元素的地址

	//cFunc.Date() 这个函数存在很大的优化空间 利用系统内的 AppendFormat()

	//go build -gcflags '-m' ./main.go 可以分析 内存逃逸 -m这个参数可以写多个，表示深层分析
}

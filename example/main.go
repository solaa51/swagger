package main

import (
	"github.com/solaa51/swagger/appServer"
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/example/controller"
	middleware2 "github.com/solaa51/swagger/example/middleware"
	"github.com/solaa51/swagger/handle/handleFuncParse"
	"github.com/solaa51/swagger/middleware"
	router "github.com/solaa51/swagger/routerV2"
)

func init() {
	//配置路由规则

	//配置全局中间件
	middleware.GlobalMiddleware = append(middleware.GlobalMiddleware, &middleware2.CrossDomain{})

	//
	r := &router.RouteParse{}
	r.BindFunc("welcome/index", func(ctx *context.Context) {
		ctx.RetData = "hello world2"
	})
	//
	//(&router.RouteParse{}).Prefix("gameApi").BindFunc("gameMall/index", func(ctx *context.Context) {
	//	ctx.RetData = "hello world2"
	//})

	(&router.RouteParse{}).Prefix("api").BindMiddleware(
		&middleware2.CheckAdmin{},
	).BindStructs(
		func() handleFuncParse.Control { return &controller.Auth{} },
	)
}

func main() {
	appServer.Run()
	// 中间件 或 路由 是否有必要支持  局部方法 限流处理。

	// 试试将segment改为没有指针的结构 这样将无GC扫描 无解啊  go限定无法获取map元素的地址
	// 放弃这个指针优化，收益与付出及其悬殊

	//go build -gcflags '-m' ./main.go 可以分析 内存逃逸 -m这个参数可以写多个，表示深层分析
}

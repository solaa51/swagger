package main

import (
	"github.com/solaa51/swagger/appServer"
	"github.com/solaa51/swagger/context"
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
}

func main() {
	appServer.Run()
	// 中间件 或 路由 是否有必要支持  局部方法 限流处理
}

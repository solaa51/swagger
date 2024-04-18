package main

import (
	"fmt"
	"github.com/solaa51/swagger/appServer"
	"github.com/solaa51/swagger/context"
	router "github.com/solaa51/swagger/routerV2"
)

func init() {
	//加载路由配置
	r := &router.RouteParse{}
	r.Prefix("api").BindFunc("welcome/index2", func(ctx *context.Context) {
		ctx.RetData = "hello world2"
	})

	(&router.RouteParse{}).Prefix("gameApi").BindFunc("gameMall/index", func(context *context.Context) {
		fmt.Println("hello game api welcome")
	})
}

func main() {
	fmt.Println("kaishi")

	appServer.Run()

	// 中间件 或 路由 是否有必要支持  局部方法 限流处理
}

func hand() {
	//handleFuncParse.ParseFunc("welcome/index", func(context *context.Context) {
	//	fmt.Println("hello world")
	//})

	//h := handler.MatchAndDelHandler("welcome/index")
	//fmt.Println(h.Call(nil))

	//h := handler.ClearHandlerToRoute()
	//fmt.Println(h["welcome/index"].Call(nil))

	//h := handler.MatchPrefixAndDelHandler("welcome")
	//fmt.Println(h)
}

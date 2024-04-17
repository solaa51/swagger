package main

import (
	"fmt"
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/example/router"
)

func main() {

	hand()

	r := &router.RouteParse{}
	r.Prefix("api").BindFunc("welcome/index2", func(context *context.Context) {
		fmt.Println("hello world2")
	})

	(&router.RouteParse{}).Prefix("gameApi").BindFunc("gameMall/index", func(context *context.Context) {
		fmt.Println("hello game api welcome")
	})

	// 给handler初始化路由
	router.InitRouterSegment()

	// 中间件 或 路由 是否有必要支持  局部方法 限流处理

	//router 根据规则匹配handler，匹配成功则删除。剩余的直接挪移到router绑定关系上
	//handler 还需要内置一个直接给router返回 map[string]*handler的方法
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

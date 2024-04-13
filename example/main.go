package main

import (
	"fmt"
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/example/router"
	"github.com/solaa51/swagger/handleFuncParse"
)

func main() {

	hand()

	r := &router.RoutParse{}
	r.Prefix("api/").BindString([]string{"1212", "456"}...).BindString("789").BindFunc("welcome/index2", func(context *context.Context) {
		fmt.Println("hello world2")
	})

	fmt.Println("路由值", router.Routers["api/welcome/index2"].String)

	//router 根据规则匹配handler，匹配成功则删除。剩余的直接挪移到router绑定关系上
	//handler 还需要内置一个直接给router返回 map[string]*handler的方法
}

func hand() {
	handleFuncParse.ParseFunc("welcome/index", func(context *context.Context) {
		fmt.Println("hello world")
	})

	//h := handler.MatchAndDelHandler("welcome/index")
	//fmt.Println(h.Call(nil))

	//h := handler.ClearHandlerToRoute()
	//fmt.Println(h["welcome/index"].Call(nil))

	//h := handler.MatchPrefixAndDelHandler("welcome")
	//fmt.Println(h)
}

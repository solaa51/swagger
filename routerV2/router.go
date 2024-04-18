package router

import (
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/handleFuncParse"
	"github.com/solaa51/swagger/log/bufWriter"
	"github.com/solaa51/swagger/middleware"
	"strings"
)

var routers = make(map[string]*Router)

// Router 外部通过Router初始化路由规则
type Router struct {
	Handler    *handleFuncParse.HandleFunc
	Middleware []middleware.Middleware
}

// RouteParse 路由解析器
type RouteParse struct {
	prefix     string                  //前缀 多次设置仅覆盖
	Middleware []middleware.Middleware //中间件 多个 可多次添加 当有一次使用后 清空
}

// 检测路由
func (r *RouteParse) checkPath(str string) string {
	s := strings.ReplaceAll(str, "//", "/")
	if s[0] == byte('/') {
		return s[1:]
	} else {
		return s
	}
}

// Prefix 设置路由前缀
func (r *RouteParse) Prefix(str string) *RouteParse {
	r.prefix = str
	return r
}

// BindFunc 绑定函数
func (r *RouteParse) BindFunc(structFuncName string, f func(*context.Context)) *RouteParse {
	if r.prefix == "" && structFuncName == "" {
		bufWriter.Info("BindFunc空路由,跳过处理")
		return r
	}

	fu := handleFuncParse.ParseFuncToRoute(structFuncName, f)
	routers[r.checkPath(r.prefix+"/"+structFuncName)] = &Router{
		Handler:    fu,
		Middleware: r.Middleware,
	}

	r.Middleware = r.Middleware[:0]

	return r
}

// BindStruct 绑定当个struct 可为struct设置别名
func (r *RouteParse) BindStruct(strut handleFuncParse.ControllerInstance, aliasName string) *RouteParse {
	ms := handleFuncParse.ParseStructToRoute(strut, aliasName)
	for k := range ms {
		routers[r.checkPath(r.prefix+"/"+k)] = &Router{
			Handler:    ms[k],
			Middleware: r.Middleware,
		}
	}

	r.Middleware = r.Middleware[:0]

	return r
}

// BindStructs 绑定多个struct
func (r *RouteParse) BindStructs(struts ...handleFuncParse.ControllerInstance) *RouteParse {
	ms := handleFuncParse.ParseStructsToRoute(struts...)
	for k := range ms {
		routers[r.checkPath(r.prefix+"/"+k)] = &Router{
			Handler:    ms[k],
			Middleware: r.Middleware,
		}
	}

	r.Middleware = r.Middleware[:0]

	return r
}

// MatchFunc 匹配hangdleFuncParse下面已解析完的方法
func (r *RouteParse) MatchFunc(newPath, structFuncName string) *RouteParse {
	if structFuncName == "" {
		return r
	}

	if r.prefix == "" && newPath == "" {
		bufWriter.Info("MatchFunc空路由,跳过处理")
		return r
	}

	f := handleFuncParse.MatchAndDelHandler(structFuncName)
	if f != nil {
		routers[r.checkPath(r.prefix+"/"+newPath)] = &Router{
			Handler:    f,
			Middleware: r.Middleware,
		}
	}

	r.Middleware = r.Middleware[:0]

	return r
}

// MatchPrefixToFunc 将包含前缀的路由 全部解析到指定的方法
func (r *RouteParse) MatchPrefixToFunc(structFuncName string) *RouteParse {
	f := handleFuncParse.MatchAndDelHandler(structFuncName)
	if f != nil {
		routers[r.checkPath(r.prefix+"/*")] = &Router{
			Handler:    f,
			Middleware: r.Middleware,
		}
	}

	r.Middleware = r.Middleware[:0]

	return r
}

// BindMiddleware 绑定中间件
func (r *RouteParse) BindMiddleware(m ...middleware.Middleware) *RouteParse {
	r.Middleware = append(r.Middleware, m...)
	return r
}

// 处理无路由规则匹配的func
func initLastHandlerFunc() {
	ms := handleFuncParse.ClearHandlerToRoute()
	for k := range ms {
		routers[k] = &Router{
			Handler: ms[k],
		}
	}
}

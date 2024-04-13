package router

import (
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/handleFuncParse"
	"github.com/solaa51/swagger/log/bufWriter"
	"github.com/solaa51/swagger/middleware"
)

var routers = make(map[string]*Router)

// Router 外部通过Router初始化路由规则
type Router struct {
	Handler    *handleFuncParse.HandleFunc
	Middleware []middleware.Middleware
	String     []string
}

// RoutParse 路由解析器
type RoutParse struct {
	prefix     string                  //前缀 多次设置仅覆盖
	Middleware []middleware.Middleware //中间件 多个 可多次添加 当有一次使用后 清空
	String     []string
}

// Prefix 设置路由前缀
func (r *RoutParse) Prefix(str string) *RoutParse {
	r.prefix = str
	return r
}

// BindFunc 绑定函数
func (r *RoutParse) BindFunc(path string, f func(*context.Context)) *RoutParse {
	if r.prefix == "" && path == "" {
		bufWriter.Info("BindFunc空路由,跳过处理")
		return r
	}

	//fmt.Println("处理前", r.String)

	fu := handleFuncParse.ParseFuncToRoute(r.prefix+path, f)
	routers[r.prefix+path] = &Router{
		Handler:    fu,
		Middleware: r.Middleware,
		String:     r.String,
	}

	r.Middleware = r.Middleware[:0]
	r.String = r.String[:0]
	//fmt.Println("处理后", r.String)

	return r
}

// BindStruct 绑定当个struct 可为struct设置别名
func (r *RoutParse) BindStruct(strut handleFuncParse.ControllerInstance, aliasName string) *RoutParse {
	ms := handleFuncParse.ParseStructToRoute(strut, aliasName)
	for k := range ms {
		routers[r.prefix+k] = &Router{
			Handler:    ms[k],
			Middleware: r.Middleware,
			String:     r.String,
		}
	}

	r.Middleware = r.Middleware[:0]
	r.String = r.String[:0]

	return r
}

// BindStructs 绑定多个struct
func (r *RoutParse) BindStructs(struts ...handleFuncParse.ControllerInstance) *RoutParse {
	ms := handleFuncParse.ParseStructsToRoute(struts...)
	for k := range ms {
		routers[r.prefix+k] = &Router{
			Handler:    ms[k],
			Middleware: r.Middleware,
			String:     r.String,
		}
	}

	r.Middleware = r.Middleware[:0]
	r.String = r.String[:0]

	return r
}

// MatchFunc 匹配hangdleFuncParse下面已解析完的方法
func (r *RoutParse) MatchFunc(newPath, structFuncName string) *RoutParse {
	if structFuncName == "" {
		return r
	}

	if r.prefix == "" && newPath == "" {
		bufWriter.Info("MatchFunc空路由,跳过处理")
		return r
	}

	f := handleFuncParse.MatchAndDelHandler(structFuncName)
	if f != nil {
		routers[r.prefix+newPath] = &Router{
			Handler:    f,
			Middleware: r.Middleware,
			String:     r.String,
		}
	}

	r.Middleware = r.Middleware[:0]
	r.String = r.String[:0]

	return r
}

// MatchPrefixToFunc 将包含前缀的路由 全部解析到指定的方法
func (r *RoutParse) MatchPrefixToFunc(structFuncName string) *RoutParse {
	f := handleFuncParse.MatchAndDelHandler(structFuncName)
	if f != nil {
		routers[r.prefix+"*"] = &Router{
			Handler:    f,
			Middleware: r.Middleware,
			String:     r.String,
		}
	}

	r.Middleware = r.Middleware[:0]
	r.String = r.String[:0]

	return r
}

// 测试使用 要删掉的
func (r *RoutParse) BindString(m ...string) *RoutParse {
	r.String = append(r.String, m...)
	return r
}

// BindMiddleware 绑定中间件
func (r *RoutParse) BindMiddleware(m ...middleware.Middleware) *RoutParse {
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

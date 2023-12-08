package handle

import (
	"encoding/json"
	"errors"
	"expvar"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"swagger/internal/appConfig"
	"swagger/internal/appPath"
	"swagger/internal/cFunc"
	"swagger/internal/context"
	"swagger/internal/control"
	"swagger/internal/limiter"
	"swagger/internal/log/bufWriter"
	"swagger/internal/middleware"
	"swagger/internal/router"
)

// http请求处理器

// HttpReturn 返回处理接口，可自定义实现该接口
type HttpReturn interface {
	End404(ctx *context.Context, err error)
	End500(ctx *context.Context, err error)
	End(ctx *context.Context)
}

type Handle struct {
	//控制器对应绑定关系
	structs map[string]control.ControllerInstance

	httpReturn HttpReturn
}

// 请求处理入口
func (h *Handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	/******特殊处理url******/
	if r.URL.Path == "/debug/var" && cFunc.LocalIP() { //公共变量的标准接口
		expvar.Handler().ServeHTTP(w, r)
		return
	}
	/************/

	// 请求日志
	bufWriter.Info("[REQUEST]",
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("ip", cFunc.ClientIP(r)),
		slog.String("user-agent", r.UserAgent()),
	)

	//调用全局中间件
	for _, m := range middleware.GlobalMiddleware {
		if !m.Handle(w, r) {
			Handler.httpReturn.End(context.NewContext(w, r, "", "", ""))
			return
		}
	}

	// 自定义路由规则匹配 解析为：class/method [params]，无法解析则返回空
	structName, methodName, args, middle := router.ParseUrlPath(r.URL.Path)
	//fmt.Println(structName, methodName, args)
	// 解析出有类和方法 进入handler匹配
	if structName != "" && methodName != "" {
		callStructName, finalStructName, has, err := checkMethod(structName, methodName, args...)
		//fmt.Println("检查方法", has, err, structName, methodName, args)
		if err != nil {
			Handler.httpReturn.End404(context.NewContext(w, r, structName, "", methodName), err)
			return
		}

		if has {
			//全局限流保护
			if limiter.Allow() {
				// 执行调用
				execCall(w, r, callStructName, finalStructName, methodName, middle, args...)
				return
			} else {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte("too many requests"))
				return
			}
		}
	}

	// 解析前端路由和文件处理
	f, err := staticFile(r.URL.Path)
	if err != nil {
		ctx := context.NewContext(w, r, "", "", "")
		Handler.httpReturn.End404(ctx, err)
		return
	}
	http.ServeFile(w, r, f)
}

// 执行方法调用
func execCall(w http.ResponseWriter, r *http.Request, structName string, finalStructName string, methodName string, middle []middleware.Middleware, args ...string) {
	//生成context
	ctx := context.NewContext(w, r, hss[structName].name, finalStructName, methodName)
	//r.WithContext(ctx.Ctx)

	var err error

	defer func() {
		if e := recover(); e != nil {
			switch e {
			default:
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)

				// 是否继续向上层抛出panic(e)
				pData, _ := json.Marshal(ctx.GetPost)
				bufWriter.Error("[REQUEST PANIC]",
					slog.String("structName", ctx.StructName),
					slog.String("methodName", ctx.MethodName),
					slog.String("requestId", ctx.RequestId),
					slog.String("method", r.Method),
					slog.String("url", r.URL.String()),
					slog.String("ip", ctx.ClientIp),
					slog.String("user-agent", r.UserAgent()),
					slog.String("paramData", string(pData)),
					slog.String("body", string(ctx.BodyData)),
					slog.String("stackInfo", string(buf[:n])),
				)

				http.Error(w, "请求处理异常", http.StatusBadGateway)
			}
		}
	}()

	// 记录请求解析后的日志
	bufWriter.Info("[REQUEST CALL]",
		slog.String("structName", ctx.StructName),
		slog.String("methodName", ctx.MethodName),
		slog.String("requestId", ctx.RequestId),
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("ip", ctx.ClientIp),
		slog.String("user-agent", r.UserAgent()),
	)

	//调用中间件处理
	for _, m := range middle {
		if !m.Handle(ctx) {
			Handler.httpReturn.End(ctx)
			return
		}
	}

	//调用方法
	err = hss[structName].call(ctx, methodName, args...)
	if err != nil {
		Handler.httpReturn.End500(context.NewContext(w, r, structName, finalStructName, methodName), err)
		return
	}

	// 统一处理返回信息
	Handler.httpReturn.End(ctx)
}

// 解析静态文件或路由转发给前端
func staticFile(urlPath string) (string, error) {
	//如果urlPath为空或首字符不是/ 则返回404
	if len(urlPath) == 0 || urlPath[0] != '/' {
		return "", errors.New("404")
	}

	// 防止恶意路由 遍历目录
	if strings.Index(urlPath, "./") >= 0 {
		return "", errors.New("404")
	}

	pp := urlPath[1:]
	if appConfig.Info().Static.Prefix != "" {
		pp = strings.Replace(pp, appConfig.Info().Static.Prefix, "", 1)
	}

	f := appPath.AppDir() + appConfig.Info().Static.LocalPath + pp

	fInfo, err := os.Stat(f)
	if err != nil {
		return "", errors.New("404")
	}

	if fInfo.IsDir() {
		//补充默认文件
		//fmt.Println("调用默认文件", f+appConfig.Info().Static.Index)
		return f + appConfig.Info().Static.Index, nil
	}

	return f, nil
}

// 检查struct 方法 参数是否匹配
func checkMethod(structName string, methodName string, args ...string) (string, string, bool, error) {
	sName := structName

	if _, ok := hss[sName]; !ok {
		sName = strings.ToUpper(sName[:1]) + sName[1:]
		if _, ok = hss[sName]; !ok {
			return "", "", false, nil
		}
	}

	v := hss[sName]
	if _, ok := v.method[methodName]; !ok {
		return "", "", false, nil
	}

	if len(v.method[methodName].in)-1 != len(args) {
		return sName, hss[sName].name, true, errors.New("参数不匹配")
	}

	return sName, hss[sName].name, true, nil
}

var Handler *Handle

func init() {
	Handler = &Handle{
		structs:    nil,
		httpReturn: defaultHttpReturn{},
	}
}

// SetCustomHttpReturn 设置自定义http返回格式
func SetCustomHttpReturn(httpReturn HttpReturn) {
	Handler.httpReturn = httpReturn
}

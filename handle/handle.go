package handle

import (
	"encoding/json"
	"errors"
	"expvar"
	"github.com/solaa51/swagger/appConfig"
	"github.com/solaa51/swagger/appPath"
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/limiter"
	"github.com/solaa51/swagger/log/bufWriter"
	"github.com/solaa51/swagger/middleware"
	"github.com/solaa51/swagger/routerV2"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// http请求处理器

// HttpReturn 返回处理接口，可自定义实现该接口
type HttpReturn interface {
	End404(ctx *context.Context, err error)
	End500(ctx *context.Context, err error)
	End(ctx *context.Context)
}

const (
	StatusZero     int = 0
	StatusOk       int = 200
	StatusNotFound int = 404
	StatusFail     int = 500
)

// 请求结束处理 记录日志
func preEnd(ctx *context.Context, status int, err error) {
	switch status {
	case StatusFail:
		Handler.httpReturn.End500(ctx, err)
	case StatusNotFound:
		Handler.httpReturn.End404(ctx, err)
	case StatusZero, StatusOk:
		Handler.httpReturn.End(ctx)
	default:
		Handler.httpReturn.End(ctx)
	}

	if ctx.Request.Method != "OPTIONS" {
		bufWriter.Info("",
			slog.Int("status", status),
			slog.String("takeTime", time.Since(ctx.StartTime).String()),
			slog.String("structFuncName", ctx.StructFuncName),
			slog.Int64("requestId", ctx.RequestId),
			slog.String("method", ctx.Request.Method),
			slog.String("url", ctx.Request.URL.String()),
			slog.String("ip", ctx.ClientIp),
			slog.String("user-agent", ctx.Request.UserAgent()),
		)
	}

	context.CtxPool.Put(ctx)
}

type Handle struct {
	httpReturn HttpReturn
}

// 请求处理入口
func (h *Handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	/******特殊处理url******/
	//if r.URL.Path == "/debug/var" && cFunc.LocalIP() { //公共变量的标准接口
	if r.URL.Path == "/debug/var" { //公共变量的标准接口
		expvar.Handler().ServeHTTP(w, r)
		return
	}
	/************/

	//调用全局中间件
	for _, m := range middleware.GlobalMiddleware {
		if !m.Handle(w, r) {
			preEnd(context.NewContext(w, r, ""), 0, nil)
			return
		}
	}

	hand, args := router.MatchHandleFunc(r.URL.Path)
	if hand != nil {
		//全局限流保护
		if limiter.Allow() {
			execCall(w, r, hand, args...)
			return
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("too many requests"))
			return
		}
	}

	// 解析前端路由和文件处理
	f, err := staticFile(r.URL.Path)
	if err != nil {
		preEnd(context.NewContext(w, r, ""), StatusNotFound, err)
		return
	}
	http.ServeFile(w, r, f)
}

// 执行方法调用
func execCall(w http.ResponseWriter, r *http.Request, handler *router.Segment, args ...string) {
	//生成context
	ctx := context.NewContext(w, r, handler.Router.Handler.StructFuncName)

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
					slog.String("structFuncName", ctx.StructFuncName),
					slog.Int64("requestId", ctx.RequestId),
					slog.String("method", r.Method),
					slog.String("url", r.URL.String()),
					slog.String("ip", ctx.ClientIp),
					slog.String("paramData", string(pData)),
					slog.String("stackInfo", string(buf[:n])),
				)

				http.Error(w, "请求处理异常", http.StatusBadGateway)
			}
		}
	}()

	//调用中间件处理
	for _, m := range handler.Router.Middleware {
		if !m.Handle(ctx) {
			preEnd(ctx, 0, nil)
			return
		}
	}

	//调用方法
	err = handler.Router.Handler.Call(ctx, args...)
	if err != nil {
		preEnd(context.NewContext(w, r, handler.Router.Handler.StructFuncName), StatusFail, err)
		return
	}

	preEnd(ctx, 0, nil)
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

var Handler = &Handle{
	httpReturn: defaultHttpReturn{},
}

// SetCustomHttpReturn 设置自定义http返回格式
func SetCustomHttpReturn(httpReturn HttpReturn) {
	Handler.httpReturn = httpReturn
}

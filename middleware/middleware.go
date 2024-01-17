package middleware

import (
	"github.com/solaa51/swagger/context"
	"net/http"
)

// Middleware 路由校验完毕后的中间件处理
type Middleware interface {
	Handle(ctx *context.Context) bool
}

// PreMiddleware 请求处理前的处理中间件 - 此时路由还未做校验
type PreMiddleware interface {
	Handle(w http.ResponseWriter, r *http.Request) bool
}

// GlobalMiddleware 挂载全局中间件处理
var GlobalMiddleware []PreMiddleware

func init() {
	GlobalMiddleware = make([]PreMiddleware, 0)
}

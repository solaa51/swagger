package control

import (
	"github.com/solaa51/swagger/context"
	"net/http/pprof"
)

type Pprof struct {
	Controller
}

func (p *Pprof) Index(ctx *context.Context) {
	pprof.Index(ctx.ResponseWriter, ctx.Request)
	ctx.CustomRet = true
}

func (p *Pprof) Cmdline(ctx *context.Context) {
	pprof.Cmdline(ctx.ResponseWriter, ctx.Request)
	ctx.CustomRet = true
}

func (p *Pprof) Profile(ctx *context.Context) {
	pprof.Profile(ctx.ResponseWriter, ctx.Request)
	ctx.CustomRet = true
}

func (p *Pprof) Symbol(ctx *context.Context) {
	pprof.Symbol(ctx.ResponseWriter, ctx.Request)
	ctx.CustomRet = true
}

func (p *Pprof) Trace(ctx *context.Context) {
	pprof.Trace(ctx.ResponseWriter, ctx.Request)
	ctx.CustomRet = true
}

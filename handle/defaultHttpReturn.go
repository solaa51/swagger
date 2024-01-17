package handle

import (
	"encoding/json"
	"net/http"
	"swagger/internal/context"
)

type defaultHttpReturn struct {
}

func (defaultHttpReturn) End404(ctx *context.Context, err error) {
	ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
	_, _ = ctx.ResponseWriter.Write([]byte(err.Error()))

	ctx.Request.Context().Done()
}

func (defaultHttpReturn) End500(ctx *context.Context, err error) {
	ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
	_, _ = ctx.ResponseWriter.Write([]byte(err.Error()))

	ctx.Request.Context().Done()
}

func (defaultHttpReturn) End(ctx *context.Context) {
	if ctx.CustomRet {
		return
	}

	if ctx.RetError != "" && ctx.RetCode == 0 {
		ctx.RetCode = 2000
	}

	if ctx.RetData == nil {
		ctx.RetData = struct{}{}
	}

	retData, _ := json.Marshal(struct {
		Msg  string `json:"msg"`
		Code int    `json:"code"`
		Data any    `json:"data"`
	}{
		ctx.RetError, ctx.RetCode, ctx.RetData,
	})

	ctx.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	ctx.ResponseWriter.Write(retData)

	ctx.Request.Context().Done()
}

package middleware

import (
	context2 "context"
	"errors"
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/example/model"
	"github.com/solaa51/swagger/library/redis"
)

// CheckAdmin 检测当前登录用户
type CheckAdmin struct {
}

func (c *CheckAdmin) Handle(ctx *context.Context) bool {
	tokenStr := ctx.Request.Header.Get("Authorization")
	if tokenStr == "" {
		ctx.RetCode = 886
		ctx.RetError = "请先登录"
		return false
	}

	if len(tokenStr) != 32 {
		ctx.RetCode = 886
		ctx.RetError = "异常token"
		return false
	}

	adminId, b, err := redis.Get(redis.KeyPrefix() + tokenStr)
	if err != nil || !b {
		ctx.RetCode = 886
		ctx.RetError = "token已失效"
		return false
	}

	//检查用户信息是否合法
	admin := &model.SysAdmin{}
	model.Db.Where("id = ? AND id_del = 0", adminId).Find(admin)
	if admin.Id == 0 {
		ctx.RetCode = 886
		ctx.RetError = "无效用户"
		return false
	}

	if admin.Status == 1 {
		ctx.RetCode = 886
		ctx.AddRetError(errors.New("您的账号已被禁用"))
		return false
	}

	//登录账号ID
	tCtx1 := context2.WithValue(ctx.Ctx, "adminId", admin.Id)
	ctx.Ctx = context2.WithValue(tCtx1, "adminRoleIds", admin.Roles)

	lastSecond, err := redis.TTL(redis.KeyPrefix() + tokenStr)
	if err != nil || lastSecond <= 0 {
		ctx.RetCode = 886
		ctx.RetError = "缓存已过期,请重新登录"
		return false
	}

	if lastSecond < 40000 {
		redis.Expire(redis.KeyPrefix()+tokenStr, 86400)
	}

	return true
}

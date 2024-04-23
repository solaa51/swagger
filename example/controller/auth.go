package controller

import (
	"errors"
	"github.com/solaa51/swagger/cFunc"
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/example/model"
	"github.com/solaa51/swagger/library/redis"
	"github.com/solaa51/swagger/snowflake"
	"strconv"
)

//登录

type Auth struct{}

func (a *Auth) Login(ctx *context.Context) {
	data, err := ctx.CheckField([]*context.CheckField{
		{Name: "username", Desc: "用户名", CheckType: context.CheckString, Request: true, Min: 5, Max: 20, Reg: "^[a-zA-Z][a-zA-Z0-9]{4,31}$"},
		{Name: "password", Desc: "密码", CheckType: context.CheckString, Request: true, Min: 4, Max: 20},
	})
	if err != nil {
		ctx.RetCode = 3000
		ctx.AddRetError(err)
		return
	}

	admin := &model.SysAdmin{}
	model.Db.Model(&model.SysAdmin{}).Where("username = ? AND is_del = 0", data["username"].(string)).Find(admin)
	if admin.Id == 0 {
		ctx.RetCode = 3008
		ctx.AddRetError(errors.New("用户名或密码错误"))
		return
	}

	if admin.Status == 1 {
		ctx.RetCode = 3008
		ctx.AddRetError(errors.New("您的账号已被禁用"))
		return
	}

	if admin.Password != md5Password(data["password"].(string)) {
		ctx.RetCode = 3008
		ctx.AddRetError(errors.New("用户名或密码错误"))

		return
	}

	//检测角色信息是否存在
	if admin.Roles == "" {
		ctx.RetCode = 3008
		ctx.AddRetError(errors.New("缺少角色信息，请找管理员处理"))
		return
	}

	model.Db.Model(&model.SysAdmin{}).Select("last_time", "last_ip").Where("id = ?", admin.Id).Updates(map[string]any{
		"last_ip":   ctx.ClientIp,
		"last_time": cFunc.Date("Y-m-d H:i:s", 0),
	})

	redisKey := cFunc.Md5([]byte(snowflake.ID() + "-" + strconv.FormatInt(admin.Id, 10)))
	err = redis.Set(redis.KeyPrefix()+redisKey, strconv.FormatInt(admin.Id, 10), 86400)
	if err != nil {
		ctx.RetCode = 3009
		ctx.AddRetError(errors.New("生成token失败，请找管理员处理"))
		return
	}

	ctx.RetData = redisKey
}

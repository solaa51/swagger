package controller

import (
	"github.com/solaa51/swagger/appConfig"
	"github.com/solaa51/swagger/cFunc"
)

// md5Password 获取密码
func md5Password(password string) string {
	str := cFunc.Md5([]byte(password)) + appConfig.Info().Md5Salt

	return cFunc.Md5([]byte(str))
}

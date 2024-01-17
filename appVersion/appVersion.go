package appVersion

import (
	"github.com/solaa51/swagger/log/bufWriter"
	"log/slog"
)

// 用来设置版本信息

// 导入 _ "goSid/internal/appVersion"
// build时设置参数
// USER 为shell id -u -n
// DATE 为shell date '+%Y-%m-%d %H:%M:%S'
// -ldflags "-X 'goSid/internal/appVersion.Version=v.1.0.0' -X 'goSid/internal/appVersion.User=$(USER)' -X 'goSid/internal/appVersion.Time=$(DATE)'"

// 可在编译时设置版本等信息
var (
	Version string
	User    string
	Time    string
)

func init() {
	if Version != "" {
		bufWriter.Info("启动版本信息：",
			slog.String("version", Version),
			slog.String("user", User),
			slog.String("time", Time),
		)
	}
}

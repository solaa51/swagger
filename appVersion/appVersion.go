package appVersion

import (
	"github.com/solaa51/swagger/log/bufWriter"
	"log/slog"
)

// 用来设置版本信息
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

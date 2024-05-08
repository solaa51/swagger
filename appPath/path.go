package appPath

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// AppDir 返回app当前的绝对目录包含目录符号
// 调试模式下如果以下未包含，可以修改调试器的输出目录
func AppDir() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	switch runtime.GOOS {
	case "windows":
		if strings.Contains(dir, "\\Temp\\go-build") {
			dir, _ = os.Getwd()
		}

		if strings.Contains(dir, "\\Temp\\GoLand") {
			dir, _ = os.Getwd()
		}
	default:
		if strings.Contains(dir, "/go-build") { //解决go run xx.go 运行模式
			dir, _ = os.Getwd()
		}

		if strings.Contains(dir, "/___go_build") || strings.Contains(dir, "/__debug") { //解决goland 和 vscode调试
			dir, _ = os.Getwd()
		}
	}

	return dir + string(os.PathSeparator)
}

// ConfigDir 返回配置文件所在的目录包含目录符号 配置文件夹为config
// 从可执行文件目录依次往上查找
// 不查找分区根目录 为了安全考虑
func ConfigDir() string {
	appDir := AppDir()

	dirSplit := strings.Split(appDir, string(os.PathSeparator))

	configDirName := "config"

	l := len(dirSplit)
	for i := l - 1; i > 1; i-- { //不允许查找分区的一级目录 为了安全考虑
		tmpDir := strings.Join(dirSplit[:i], string(os.PathSeparator))
		if _, err := os.Stat(tmpDir + string(os.PathSeparator) + configDirName); err == nil {
			return tmpDir + string(os.PathSeparator) + configDirName + string(os.PathSeparator)
		}
	}

	err := os.Mkdir(appDir+configDirName, 0755)
	if err != nil {
		log.Fatal("创建配置文件目录失败：", appDir+configDirName, err.Error())
	}

	return appDir + configDirName + string(os.PathSeparator)
}

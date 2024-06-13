package configFiles

import (
	"embed"
	"errors"
	"github.com/solaa51/swagger/appPath"
	"io/fs"
	"os"
)

var ConfigDir embed.FS

func SetEmbedConfigDir(dir embed.FS) {
	ConfigDir = dir
}

func GetConfigFile(fileName string) ([]byte, error) {
	if _, err := os.Stat(appPath.ConfigDir() + fileName); err == nil {
		return os.ReadFile(appPath.ConfigDir() + fileName)
	}

	if os.Getenv("EMBED-CONFIG-FILES") == "1" {
		return fs.ReadFile(ConfigDir, "config/"+fileName)
	}

	return nil, errors.New("not found")
}

func GetConfigPath(fileName string) string {
	return appPath.ConfigDir() + fileName
}

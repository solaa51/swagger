package appConfig

import (
	"github.com/solaa51/swagger/appPath"
	"github.com/solaa51/swagger/cFunc"
	"github.com/solaa51/swagger/configFiles"
	"github.com/solaa51/swagger/log/bufWriter"
	"github.com/solaa51/swagger/watchConfig"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

/**
配置文件解析 app.toml
*/

var config = &Config{}

// Http http服务配置
type Http struct {
	PORT     string `yaml:"port"`  //http 监听端口
	HTTPS    bool   `yaml:"https"` //是否开启https服务
	HTTPSKEY string `yaml:"httpsKey"`
	HTTPSPEM string `yaml:"httpsPem"`
}

// StaticConfig 静态文件及路由匹配配置
type StaticConfig struct {
	Prefix    string `yaml:"prefix"`    //html js等代码内的前缀路径 默认assets/
	LocalPath string `yaml:"localPath"` //本地存储的真实目录 默认front_end/
	Index     string `yaml:"index"`     //前端入口文件 默认index.html
}

// RateConfig 全局限流配置
type RateConfig struct {
	PerSecond       float64 `yaml:"perSecond"`       //每秒流入桶内的数量
	Bucket          int     `yaml:"bucket"`          // 桶容量
	WaitMillisecond int     `yaml:"waitMillisecond"` //允许等待的超时毫秒数
}

type Config struct {
	Http Http `yaml:"http"`

	// 静态目录配置
	Static StaticConfig `yaml:"staticDir"`

	//本地local 测试test 预发pre 生产prod
	//默认为local 其他值则日志不会在标准打印
	Env string `yaml:"env"`

	//md5加密盐
	Md5Salt string `yaml:"md5Salt"`

	// 全局限流配置
	Rate RateConfig `yaml:"rateConfig"`

	//服务实例节点ID
	ServerId int64 `yaml:"serverId"`
}

// Info 配置信息
func Info() *Config {
	return config
}

// 检查http配置
func (c *Config) checkHttpConfig() {
	if config.Http.PORT == "" { //初始 服务刚启动 未配置端口
		if c.Http.PORT == "" {
			c.Http.PORT, _ = cFunc.GetFreePort()
		}

		return
	}

	if c.Http.PORT == "" { //沿用当前配置
		c.Http.PORT = config.Http.PORT
		c.Http.HTTPS = config.Http.HTTPS
		c.Http.HTTPSKEY = config.Http.HTTPSKEY
		c.Http.HTTPSPEM = config.Http.HTTPSPEM
		return
	}

	if c.Http.PORT != config.Http.PORT {
		bufWriter.Warn("app.yaml配置http端口与当前配置不一致，热更新无法修改端口")
		c.Http.PORT = config.Http.PORT
	}

	if c.Http.HTTPS != config.Http.HTTPS ||
		c.Http.HTTPSKEY != config.Http.HTTPSKEY ||
		c.Http.HTTPSPEM != config.Http.HTTPSPEM {

		execFile, _ := filepath.Abs(os.Args[0])
		bufWriter.Warn(execFile, "app.yaml文件下http服务配置变更触发重启更新，发送热更新信号")
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGHUP)
	}
}

func (c *Config) check() {
	//初始化部分配置信息
	if c.Env == "" {
		c.Env = "local"
	}

	if c.ServerId == 0 {
		c.ServerId = 1
	}

	//if c.Static.Prefix == "" {
	//	c.Static.Prefix = "assets/"
	//}
	if c.Static.LocalPath == "" {
		c.Static.LocalPath = "front_end/"
	}
	if c.Static.Index == "" {
		c.Static.Index = "index.html"
	}

	if strings.Contains(c.Static.LocalPath, "./") {
		bufWriter.Fatal("目录地址不允许出现./字符", c.Static.LocalPath)
	}

	if c.Http.HTTPS {
		if c.Http.HTTPSPEM == "" || c.Http.HTTPSKEY == "" {
			bufWriter.Fatal("请为https服务配置证书:httpsKey和httpsPem")
		}

		if _, err := configFiles.GetConfigFile(c.Http.HTTPSKEY); err != nil {
			bufWriter.Fatal("请为https服务配置证书:httpsKey")
		}

		if _, err := configFiles.GetConfigFile(c.Http.HTTPSPEM); err != nil {
			bufWriter.Fatal("请为https服务配置证书:httpsPem")
		}

		// 兼容embed后 这个地址就无效了
		//c.Http.HTTPSKEY = appPath.ConfigDir() + c.Http.HTTPSKEY
		//c.Http.HTTPSPEM = appPath.ConfigDir() + c.Http.HTTPSPEM
	}
}

func parseConfigFile() (*Config, error) {
	c := &Config{}
	f, err := configFiles.GetConfigFile("app.yaml")
	if err != nil {
		bufWriter.Warn("未配置" + appPath.ConfigDir() + "app.yaml")
	} else {
		err = yaml.Unmarshal(f, c)
		if err != nil {
			bufWriter.Warn("解析app.yaml配置文件失败", err)
			return nil, err
		}
	}

	return c, nil
}

func newConfig() *Config {
	c, err := parseConfigFile()
	if err != nil {
		c = &Config{}
	}

	c.check()
	c.checkHttpConfig()

	if c.Env == "" || c.Env == "local" {
		bufWriter.SetDefaultStdout(true)
	} else {
		bufWriter.SetDefaultStdout(false)
	}
	bufWriter.SetDefaultLevel(c.Env)

	return c
}

func init() {
	config = newConfig()

	ch, _ := watchConfig.AddWatch(configFiles.GetConfigPath("app.yaml"))
	go func() {
		for {
			select {
			case <-ch:
				bufWriter.Info(appPath.ConfigDir()+"app.yaml", "文件变更触发更新")
				config = newConfig()
			}
		}
	}()
}

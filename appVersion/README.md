启动后输出版本信息
编译时写入版本信息

使用方法：
入口导入
_ "goSid/internal/appVersion"

    编译命令举例
        go build -ldflags "-X 'goSid/internal/appVersion.Version=v.1.0.0' -X 'goSid/internal/appVersion.User=$(id -u -n)' -X 'goSid/internal/appVersion.Time=$(date +%Y-%m-%d+%H:%M:%S)' -s -w" -o b b.go
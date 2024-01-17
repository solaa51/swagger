启动后输出版本信息
编译时写入版本信息

使用方法：
入口导入
_ "github.com/solaa51/swagger/appVersion"

    编译命令举例
        go build -ldflags "-X 'github.com/solaa51/swagger/appVersion.Version=v.1.0.0' -X 'github.com/solaa51/swagger/appVersion.User=$(id -u -n)' -X 'github.com/solaa51/swagger/appVersion.Time=$(date +%Y-%m-%d+%H:%M:%S)' -s -w" -o b b.go

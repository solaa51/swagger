package app

import (
	"os"
	"os/signal"
	"syscall"
)

type app struct {
	close   []func() //回收资源
	restart []func() //重启
}

// RegistClose 注册回收资源
func RegistClose(f func()) {
	a.close = append(a.close, f)
}

// RegistRestart 注册重启资源
func RegistRestart(f func()) {
	a.restart = append(a.restart, f)
}

// Close 当不使用信号处理时，可直接调用Close回收资源
func Close() {
	a.closeFunc()
}

func (a *app) closeFunc() {
	for _, f := range a.close {
		f()
	}
}

func (a *app) restartFunc() {
	for _, f := range a.restart {
		f()
	}
}

// ListenSignal 监听系统信号
func ListenSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch,
		syscall.SIGHUP, //用于更新或重启 kill -1 / kill -HUP
		syscall.SIGINT, //kill -2 软退出/ctrl+c
		//syscall.SIGKILL, //kill -9 强制关闭进程 无法捕获
		syscall.SIGQUIT, //
		syscall.SIGTERM, //kill -15 优雅地终止进程 docker的stop会优先发送该信号，若一直不退出则发送kill -9信号
	)

	for {
		switch <-ch {
		case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM: // 关闭信号
			//fmt.Println("关闭信号")

			signal.Stop(ch) //关闭信号通道
			a.closeFunc()

			break
		case syscall.SIGHUP: //自定义的更新重启信号
			//fmt.Println("重启信号")
			signal.Stop(ch) //关闭信号通道
			a.restartFunc()
			a.closeFunc()

			break
		default:
			continue
		}
	}
}

var a = &app{}

package http

import (
	"context"
	"errors"
	"flag"
	"github.com/solaa51/swagger/appConfig"
	"github.com/solaa51/swagger/handle"
	"github.com/solaa51/swagger/log/bufWriter"
	"github.com/solaa51/swagger/watchConfig"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type shutdownSingle struct {
	server   *http.Server //http服务server配置
	listener net.Listener
}

// 监听信号
func (s *shutdownSingle) listenSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	switch <-ch {
	case syscall.SIGINT, syscall.SIGTERM: // 关闭信号
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		signal.Stop(ch) //关闭信号通道

		err := s.server.Shutdown(ctx) //平滑关闭连接中的请求
		if err != nil {
			bufWriter.Error("关闭服务失败:", err)
		}

		//关闭日志缓冲区
		bufWriter.CloseDefault()

		return
	case syscall.SIGHUP: //自定义的更新重启信号
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := s.restart()
		if err != nil {
			bufWriter.Error("热更新服务失败:", err)
			return
		}

		bufWriter.Info("旧服务开始关闭", os.Getpid())
		err = s.server.Shutdown(ctx) //平滑关闭连接中的请求
		if err != nil {
			bufWriter.Error("关闭服务失败:", err)
		}
		bufWriter.Info("旧服务关闭成功", os.Getpid())

		//返回并退出
		return
	}
}

// 重启服务 启用新进程接收新的请求
func (s *shutdownSingle) restart() error {
	bufWriter.Info("开始重启服务")

	ln := s.listener.(*net.TCPListener)

	ff, err := ln.File()
	if err != nil {
		return errors.New("获取socket文件描述符失败")
	}

	cmd := exec.Command(os.Args[0], []string{"-g"}...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{ff} //重用原有的socket文件描述符

	err = cmd.Start()
	if err != nil {
		return errors.New("启动新进程失败:" + err.Error())
	}

	return nil
}

// 监控启动文件的变更
func (s *shutdownSingle) watchSelf() {
	execFile, _ := filepath.Abs(os.Args[0])
	ch, _ := watchConfig.AddWatch(execFile)
	go func() {
		for {
			select {
			case <-ch:
				f, _ := os.Stat(execFile)
				perm := f.Mode().Perm().String()
				if strings.Count(perm, "x") < 3 {
					err := os.Chmod(execFile, 0755)
					if err != nil {
						bufWriter.Error(execFile, "修改可执行权限失败", err)
						continue
					}
				}

				bufWriter.Info(execFile, "文件变更触发重启更新，发送热更新信号")
				p, _ := os.FindProcess(os.Getpid())
				_ = p.Signal(syscall.SIGHUP)
			}
		}
	}()
}

func Run() {
	var (
		d bool
		g bool
	)
	flag.BoolVar(&d, "d", false, "后台执行")
	flag.BoolVar(&g, "g", false, "平滑重启，不需要手动调用")

	flag.Parse()

	daemon(d)
	start(g)
}

func start(restart bool) {
	var ln net.Listener
	var httpAddr string
	var err error

	if restart {
		//启动命令中包含参数 热重启时，从socket文件描述符 重新启动一个监听
		//当存在监听socket时 socket的文件描述符就是3 所以从本进程的3号文件描述符 恢复socket监听
		f := os.NewFile(3, "")
		ln, err = net.FileListener(f)
		if err != nil {
			bufWriter.Fatal("重启服务失败", err)
		}

		addr := ln.Addr().(*net.TCPAddr)
		httpAddr = ":" + strconv.Itoa(addr.Port)

		bufWriter.Info("升级重启", os.Args, httpAddr, os.Getpid())
	} else {
		httpAddr = getHttpAddr()
		ln, err = net.Listen("tcp", httpAddr)
		if err != nil {
			bufWriter.Fatal("监听端口失败", err)
		}
	}

	server := &http.Server{
		Addr:    httpAddr,
		Handler: handle.Handler,
	}

	server.RegisterOnShutdown(func() {
		bufWriter.SetDefaultBuffer(false) //关闭日志缓冲区
		bufWriter.CloseDefault()          //关闭日志文件句柄
	})

	// 启动http服务监听
	go func() {
		if appConfig.Info().Http.HTTPS {
			err = server.ServeTLS(ln, appConfig.Info().Http.HTTPSPEM, appConfig.Info().Http.HTTPSKEY)
		} else {
			err = server.Serve(ln)
		}

		bufWriter.Fatal("服务启动失败或已被关闭", err, server.Addr, os.Getpid())
	}()

	bufWriter.Info("启动服务，监听地址为:", server.Addr, "进程ID为:", os.Getpid())

	ss := &shutdownSingle{
		server:   server,
		listener: ln,
	}

	go ss.watchSelf()

	//设置默认日志设置
	defaultLogSet()

	ss.listenSignal()
}

// 设置默认日志设置
func defaultLogSet() {
	bufWriter.SetDefaultBuffer(true) //开启缓冲区

	bufWriter.SetDefaultLevel(appConfig.Info().Env)
}

// 获取http服务启动所需的监听端口号
func getHttpAddr() string {
	// 获取配置的http监听端口，默认随机分配
	addr := ":" + appConfig.Info().Http.PORT

	return addr
}

// 进入守护进程
func daemon(d bool) {
	if d && os.Getppid() != 1 { //判断父进程  父进程为1则表示已被系统接管
		filePath, _ := filepath.Abs(os.Args[0]) //将启动命令 转换为 绝对地址命令
		cmd := exec.Command(filePath, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Start()

		os.Exit(0)
	}
}

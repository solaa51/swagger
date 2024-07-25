package appServer

import (
	"context"
	"crypto/tls"
	"flag"
	"github.com/solaa51/swagger/app"
	"github.com/solaa51/swagger/appConfig"
	"github.com/solaa51/swagger/appVersion"
	"github.com/solaa51/swagger/cFunc"
	"github.com/solaa51/swagger/configFiles"
	"github.com/solaa51/swagger/handle"
	"github.com/solaa51/swagger/log/bufWriter"
	router "github.com/solaa51/swagger/routerV2"
	"github.com/solaa51/swagger/watchConfig"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
)

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

	app.ListenSignal()
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

	//加载路由
	router.InitRouterSegment()

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
			//兼容embed后 这里不能使用地址，需要改为直接读取内容
			var config *tls.Config
			if server.TLSConfig == nil {
				config = &tls.Config{}
			} else {
				config = server.TLSConfig.Clone()
			}
			if !slices.Contains(config.NextProtos, "http/1.1") {
				config.NextProtos = append(config.NextProtos, "http/1.1")
			}

			config.Certificates = make([]tls.Certificate, 1)
			certFile, _ := configFiles.GetConfigFile(appConfig.Info().Http.HTTPSPEM)
			keyFile, _ := configFiles.GetConfigFile(appConfig.Info().Http.HTTPSKEY)
			config.Certificates[0], err = tls.X509KeyPair(certFile, keyFile)
			if err != nil {
				bufWriter.Fatal("证书文件解析失败", err)
			}

			ln = tls.NewListener(ln, config)

			//err = server.ServeTLS(ln, appConfig.Info().Http.HTTPSPEM, appConfig.Info().Http.HTTPSKEY)
		}

		err = server.Serve(ln)

		bufWriter.Fatal("服务启动失败或已被关闭", err, server.Addr, os.Getpid())
	}()

	bufWriter.Warn("启动服务，监听地址:", server.Addr, "进程ID:", os.Getpid(), "版本号：", appVersion.Version)

	ss := &appServer{
		server:   server,
		listener: ln,
	}

	//设置默认日志设置
	defaultLogSet()

	go ss.watchSelf()

	app.RegistClose(ss.shutdown)
	app.RegistRestart(ss.restart)
}

type appServer struct {
	server   *http.Server //http服务server配置
	listener net.Listener
}

// 重启服务 启用新进程接收新的请求
func (s *appServer) restart() {
	bufWriter.Warn("开始重启服务")

	ln := s.listener.(*net.TCPListener)

	ff, err := ln.File()
	if err != nil {
		bufWriter.Error("获取socket文件描述符失败", err)
		return
	}

	cmd := exec.Command(os.Args[0], []string{"-g"}...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{ff} //重用原有的socket文件描述符

	err = cmd.Start()
	if err != nil {
		bufWriter.Error("重启启动新进程失败:" + err.Error())
		return
	}
}

// 平滑关闭连接
func (s *appServer) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.server.Shutdown(ctx) //平滑关闭连接中的请求
	if err != nil {
		bufWriter.Error("关闭服务失败:", err)
	}
}

// 监控启动文件的变更
func (s *appServer) watchSelf() {
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

				bufWriter.Warn(execFile, "文件变更触发重启更新，发送热更新信号")
				p, _ := os.FindProcess(os.Getpid())
				_ = p.Signal(syscall.SIGHUP)
			}
		}
	}()
}

// 获取http服务启动所需的监听端口号
func getHttpAddr() string {
	// 获取配置的http监听端口，默认随机分配
	addr := ":" + appConfig.Info().Http.PORT

	return addr
}

// 设置默认日志设置
func defaultLogSet() {
	bufWriter.SetDefaultBuffer(true) //开启缓冲区

	bufWriter.SetDefaultLevel(appConfig.Info().Env)
}

// 进入守护进程
func daemon(d bool) {
	if d {
		cFunc.Daemon()
	}
}

package pprofStack

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"syscall"
)

// Pprof 分析采样
func Pprof() {
	//自定义信号识别码
	sig := syscall.Signal(31)
	ch := make(chan os.Signal)

	flag := false

	//监听信号
	signal.Notify(ch, sig)

	go func() {
		var cpuProfile, memoryProfile, traceProfile *os.File
		for range ch {
			if !flag {
				//开始采集
				cpuProfile, _ := os.Create("cpu.pprof")
				memoryProfile, _ := os.Create("memory.pprof")
				traceProfile, _ := os.Create("runtime.trace")
				_ = pprof.StartCPUProfile(cpuProfile)
				_ = pprof.WriteHeapProfile(memoryProfile)
				_ = trace.Start(traceProfile)

				flag = true
			} else {
				//结束
				pprof.StopCPUProfile()
				trace.Stop()
				memoryProfile.Close()
				cpuProfile.Close()
				traceProfile.Close()
				flag = false
			}
		}
	}()
}

// Stack 打印堆栈信息
func Stack() {
	//自定义信号识别码
	sig := syscall.Signal(32)
	ch := make(chan os.Signal)
	signal.Notify(ch, sig)
	go func() {
		for range ch {
			buffer := make([]byte, 1024*1024*4)
			runtime.Stack(buffer, true)
			fmt.Println(string(buffer))
		}
	}()
}

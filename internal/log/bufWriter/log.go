package bufWriter

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// 可暴露 1. 前缀  2. 缓冲区的启用 动态修改
// 可操作 3 日志级别
// 可操作 env改变 修改存储路径 可直接读取env多判断一次 写入到每个级别上
// 日志需要优先于config等的加载

// 默认日志处理为直接写入文件，当env==local时，日志会打印在终端
// 所有检查全部通过后，再调用打开缓冲区
var dl *defaultLog

type defaultLog struct {
	writer *BufWriter
	logger *slog.Logger   // 日志记录器
	level  *slog.LevelVar // 日志级别
	//buffer bool           // 是否启用缓冲区写入日志
	prefix string // 日志文件前缀
}

// SetDefaultLevel 动态修改默认日志级别
func SetDefaultLevel(env string) {
	var level slog.Level
	switch env {
	case "test":
		level = slog.LevelInfo
	case "pre":
		level = slog.LevelWarn
	case "prod":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	dl.level.Set(level)
}

// SetDefaultStdout 修改终端是否输出日志
func SetDefaultStdout(stdout bool) {
	dl.writer.stdout = stdout
}

// SetDefaultBuffer 动态修改buffer缓冲写入
func SetDefaultBuffer(buffer bool) {
	dl.writer.buffer = buffer
}

// CloseDefault 回收资源
func CloseDefault() {
	dl.writer.Close()
}

func Info(msg string, args ...any) {
	if len(args) > 0 {
		dl.logger.Info(msg + fmt.Sprint(args...))
	} else {
		dl.logger.Info(msg)
	}
}

func Error(msg string, args ...any) {
	if len(args) > 0 {
		dl.logger.Error(msg+fmt.Sprint(args...), slog.Any("source", caller()))
	} else {
		dl.logger.Error(msg, slog.Any("source", caller()))
	}
}

func Warn(msg string, args ...any) {
	if len(args) > 0 {
		dl.logger.Warn(msg + fmt.Sprint(args...))
	} else {
		dl.logger.Warn(msg)
	}
}

func Fatal(msg string, args ...any) {
	Error(msg, args)

	os.Exit(1)
}

func caller() *slog.Source {
	var source *slog.Source

	pc, file, _, ok := runtime.Caller(3)
	if ok {
		file = filepath.Base(file)
		fs := runtime.CallersFrames([]uintptr{pc})
		f, _ := fs.Next()
		source = &slog.Source{
			Function: f.Function,
			File:     f.File,
			Line:     f.Line,
		}
	}

	/*// 调试获取调用文件的层级
	for i := 1; i < 25; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if ok {
			file = filepath.Base(file)
			funcName := runtime.FuncForPC(pc).Name()
			fmt.Println(i, "-", file, "--", funcName, "---", line)
		}
	}*/

	return source
}

func init() {
	dl = &defaultLog{
		level:  &slog.LevelVar{},
		prefix: "log-",
	}
	dl.level.Set(slog.LevelInfo)

	options := &slog.HandlerOptions{
		AddSource: false, //标识打印日志的来源文件信息 可自定义实现该功能 错误以上的加上调用位置即可
		Level:     dl.level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
			}

			if a.Key == slog.TimeKey {
				return slog.String(a.Key, a.Value.Time().Format(time.DateTime+".000000"))
			}

			return a
		},
	}

	w := NewBufWriter(dl.prefix, false, true)
	dl.writer = w

	dl.logger = slog.New(slog.NewJSONHandler(w, options))
}

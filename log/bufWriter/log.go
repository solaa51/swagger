package bufWriter

import (
	"fmt"
	"github.com/solaa51/swagger/app"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// SwaLog 日志处理器
type SwaLog struct {
	writer *BufWriter
	logger *slog.Logger   // 日志记录器
	level  *slog.LevelVar // 日志级别
}

// SetLevel 修改日志级别
func (s *SwaLog) SetLevel(env string) {
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

	s.level.Set(level)
}

// SetBuffer 动态修改buffer缓冲区开关
func (s *SwaLog) SetBuffer(buffer bool) {
	s.writer.buffer = buffer
}

// SetStdout 修改终端是否输出日志
func (s *SwaLog) SetStdout(stdout bool) {
	s.writer.stdout = stdout
}

// Close 回收资源
func (s *SwaLog) Close() {
	s.writer.Close()
}

func (s *SwaLog) Info(msg string, args ...any) {
	if len(args) > 0 {
		s.logger.Info(msg + fmt.Sprint(args...))
	} else {
		s.logger.Info(msg)
	}
}

func (s *SwaLog) Error(msg string, args ...any) {
	if len(args) > 0 {
		s.logger.Error(msg+fmt.Sprint(args...), slog.Any("source", caller()))
	} else {
		s.logger.Error(msg, slog.Any("source", caller()))
	}
}

func (s *SwaLog) Warn(msg string, args ...any) {
	if len(args) > 0 {
		s.logger.Warn(msg + fmt.Sprint(args...))
	} else {
		s.logger.Warn(msg)
	}
}

func (s *SwaLog) Fatal(msg string, args ...any) {
	s.SetBuffer(false)
	s.Error(msg, args)
	os.Exit(1)
}

func NewSwaLog(prefix string, buffer bool, stdout bool) *SwaLog {
	ll := &SwaLog{
		level: &slog.LevelVar{},
	}
	ll.level.Set(slog.LevelInfo)

	options := &slog.HandlerOptions{
		Level: ll.level,
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

	w := NewBufWriter(prefix, buffer, stdout)
	ll.writer = w

	ll.logger = slog.New(slog.NewJSONHandler(w, options))

	return ll
}

var defaultLog *SwaLog

// SetDefaultLevel 动态修改默认日志级别
func SetDefaultLevel(env string) {
	defaultLog.SetLevel(env)
}

// SetDefaultStdout 修改终端是否输出日志
func SetDefaultStdout(stdout bool) {
	defaultLog.SetStdout(stdout)
}

// SetDefaultBuffer 动态修改buffer缓冲写入
func SetDefaultBuffer(buffer bool) {
	defaultLog.SetBuffer(buffer)
}

// CloseDefault 回收资源
func CloseDefault() {
	defaultLog.Close()
}

func Info(msg string, args ...any) {
	defaultLog.Info(msg, args)
}

func Error(msg string, args ...any) {
	defaultLog.Error(msg, args)
}

func Warn(msg string, args ...any) {
	defaultLog.Warn(msg, args)
}

func Fatal(msg string, args ...any) {
	defaultLog.Fatal(msg, args)
}

func caller() *slog.Source {
	var source *slog.Source

	pc, file, line, ok := runtime.Caller(3)
	if ok {
		file = filepath.Base(file)
		funcName := runtime.FuncForPC(pc).Name()
		source = &slog.Source{
			Function: funcName,
			File:     file,
			Line:     line,
		}
	}

	// 调试获取调用文件的层级
	/*for i := 1; i < 25; i++ {
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
	defaultLog = NewSwaLog("log-", false, true)
	app.RegistClose(defaultLog.Close)
}

package orm

import (
	"fmt"
	"strings"
	"swagger/internal/cFunc"
	"swagger/internal/log/bufWriter"
)

func newDbLogWriter() *fileWriter {
	f := &fileWriter{
		writer: bufWriter.NewBufWriter("gorm-", false, false),
	}
	return f
}

// sql日志 文件输出
type fileWriter struct {
	writer *bufWriter.BufWriter
}

func (f *fileWriter) Printf(format string, v ...any) {
	ff := strings.ReplaceAll(format, "\n", "\n\t")
	logStr := fmt.Sprintf(ff, v...)
	_, _ = f.writer.Write([]byte("\n[" + cFunc.Date("Y-m-d H:i:s", 0) + "] " + logStr))
}

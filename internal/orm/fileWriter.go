package orm

import (
	"fmt"
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
	logStr := fmt.Sprintf(format, v...)
	_, _ = f.writer.Write([]byte(logStr))
}

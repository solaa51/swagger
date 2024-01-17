package orm

import (
	"fmt"
	"github.com/solaa51/swagger/cFunc"
	"github.com/solaa51/swagger/log/bufWriter"
	"strings"
)

// prefix 数据库日志文件前缀
func newDbLogWriter(prefix string) *fileWriter {
	f := &fileWriter{
		writer: bufWriter.NewBufWriter(prefix, false, false),
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

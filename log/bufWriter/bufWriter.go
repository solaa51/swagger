package bufWriter

import (
	"bufio"
	"fmt"
	"github.com/solaa51/swagger/appPath"
	"github.com/solaa51/swagger/cFunc"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

func NewBufWriter(prefix string, buffer bool, stdout bool) *BufWriter {
	w := &BufWriter{
		prefix:   prefix,
		suffix:   ".log",
		path:     appPath.AppDir() + "logs" + string(os.PathSeparator),
		saveDays: 14,
		buffer:   buffer,
		stdout:   stdout,
	}

	w.init()

	return w
}

type BufWriter struct {
	logFile  *os.File
	writer   *bufio.Writer
	buffer   bool //是否启用缓冲区
	stdout   bool //是否在终端输出
	wg       sync.Mutex
	close    chan struct{}
	curDate  string //用来处理日志分片
	prefix   string //日志文件名前缀
	suffix   string //日志文件后缀
	path     string //日志存储目录
	saveDays int    //日志保留天数
}

func (w *BufWriter) createFile() {
	if _, err := os.Stat(w.path); err != nil {
		_ = os.MkdirAll(w.path, os.ModePerm)
	}

	w.curDate = cFunc.Date("Y-m-d", 0)
	fileName := w.path + w.prefix + w.curDate + w.suffix
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		log.Fatal("无日志写入权限")
	}

	writer := bufio.NewWriter(logFile)

	w.logFile = logFile
	w.writer = writer

	// 清理过期日志文件
	fs, _ := os.ReadDir(w.path)
	for _, v := range fs {
		if v.IsDir() {
			continue
		}

		if strings.HasPrefix(v.Name(), w.prefix) && strings.HasSuffix(v.Name(), w.suffix) {
			f, _ := v.Info()
			if f.ModTime().Before(time.Now().AddDate(0, 0, w.saveDays*-1)) {
				_ = os.Remove(w.path + v.Name())
			}
		}
	}
}

// 初始化对象
func (w *BufWriter) init() {
	w.createFile()
	w.close = make(chan struct{})

	//if w.buffer {
	go func() {
		timer := time.NewTimer(time.Millisecond * 100)
		for {
			select {
			case <-timer.C:
				w.flush()
				timer.Reset(time.Millisecond * 100)
			case <-w.close:
				timer.Stop()

				return
			}
		}
	}()
	//}
}

func (w *BufWriter) Write(p []byte) (n int, err error) {
	w.wg.Lock()
	defer w.wg.Unlock()

	nData := cFunc.Date("Y-m-d", 0)
	if nData != w.curDate {
		_ = w.writer.Flush()
		_ = w.logFile.Close()

		w.createFile()
	}

	if w.stdout {
		//fmt.Println("终端输出")
		fmt.Print(string(p))
	}

	if w.buffer {
		//fmt.Println("缓冲区写入文件")
		return w.writer.Write(p)
	} else {
		//fmt.Println("直接写入文件")
		return w.logFile.Write(p)
	}
}

func (w *BufWriter) flush() {
	if w.writer.Size() > 0 {
		w.wg.Lock()
		defer w.wg.Unlock()
		_ = w.writer.Flush()
	}
}

func (w *BufWriter) Close() {
	w.wg.Lock()
	defer w.wg.Unlock()

	_ = w.writer.Flush()
	_ = w.logFile.Close()

	if w.buffer {
		w.close <- struct{}{}
	}
}

package watchConfig

import (
	"github.com/solaa51/swagger/cFunc"
	"github.com/solaa51/swagger/log/bufWriter"
	"os"
	"time"
)

type watchFile struct {
	md5     string
	modTime int64 //文件修改时间
	ch      []chan struct{}
}

var watchFiles map[string]*watchFile

// AddWatch 添加监控文件 返回channel便于发送通知
func AddWatch(filePath string) (chan struct{}, error) {
	n := make(chan struct{})
	if _, ok := watchFiles[filePath]; !ok {
		var modTime int64
		var fileMd5 string
		if f, err := os.Stat(filePath); err != nil {
			fileMd5 = ""
			modTime = 0
		} else {
			fileMd5, _ = cFunc.Md5File(filePath)
			modTime = f.ModTime().Unix()
		}

		watchFiles[filePath] = &watchFile{
			modTime: modTime,
			md5:     fileMd5,
			ch:      make([]chan struct{}, 0),
		}
	}

	watchFiles[filePath].ch = append(watchFiles[filePath].ch, n)

	return n, nil
}

func init() {
	watchFiles = make(map[string]*watchFile)

	go func() {
		t := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-t.C:
				for k, v := range watchFiles {
					if f, err := os.Stat(k); err != nil {
						//记录错误日志
						//bufWriter.Error("文件变更监听错误："+k+" ", err)
						continue
					} else {
						//先检查文件变更时间
						if f.ModTime().Unix() != v.modTime {
							v.modTime = f.ModTime().Unix()

							fileMd5, err := cFunc.Md5File(k)
							if err != nil {
								//记录错误日志
								bufWriter.Error("文件变更监听错误："+k+" ", err)
								continue
							}

							//再检查文件md5值变更
							if fileMd5 != v.md5 {
								v.md5 = fileMd5
								//发送消息
								//fmt.Println("发送变更通知")
								for _, c := range v.ch {
									c <- struct{}{}
								}
							}
						}
					}
				}
			}
		}
	}()
}

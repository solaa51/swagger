## 利用读取文件的修改时间来处理变动事件

    该变更利用了Linux系统下变更文件名为原子操作来保证了安全

```
n, _ := addWatch(configDir + "app.toml")
for {
    select {
    case <-n:
        fmt.Println("收到了变更通知")
    }
}
```

执行cmd命令
执行ssh命令
利用sftp传输文件
scp其实就是遍历文件，利用sftp协议

以后可扩展出scp的功能，直接传输或下载文件夹

```
localFile := appPath.AppDir() + "release/keluoou-v2.2.2/keluoou3"
client, err := cmdRun.NewSSH("118.195.161.65:22")
if err != nil {
    fmt.Println(err)
return
}

client.Run("ls -alx")

dstFile := "/data/keluoou"
err = client.SftpSend(localFile, dstFile)
if err != nil {
    fmt.Println(err)
return
}
```

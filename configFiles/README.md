读取配置文件

    支持go embed
        可将配置文件嵌入到可执行文件中
    本地配置文件优先级高于内嵌文件

如果使用go embed
需在入口main包嵌入以下代码

```
//go:embed config
var embedConfigFiles embed.FS

func init() {
	_ = os.Setenv("EMBED-CONFIG-FILES", "1")
	tools.SetEmbedConfigDir(embedConfigFiles)
}
```

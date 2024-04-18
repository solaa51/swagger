# swagger
go-web框架

## 待完善计划表

## go标准库rpc性能大概有grpc的2倍

## 路由处理 - 改为链表 匹配性能提升100倍 整体提升没多大[100µs]
    将正则规则转换为字符串处理
    服务启动时 处理完所有路由规则 并将其注册到服务中
    利用链表存储路由规则

    正则匹配下的耗时36.827µs
    链表匹配下的耗时500ns

    handler 先生成默认的路由映射规则
        handler 可追加内置处理 比如pprof
    router 再对规则二次定义
        初始化后对handler做删除，使handler全部移交给router处理

## 对外开放的方法及路由规则 调整为显式加载 减少隐式加载的黑洞

## http pprof内置支持
    
    可自定义重写路由规则 方便追加验证规则 加强安全性

## /debug/var内置支持

    支持自定义路由规则

## 版本信息 是否也能默认加载 减少隐式加载的黑洞

## json参数校验规则 是否能继续加强。
    大部分post提交的参数都不会很复杂，也不太多。解析字符串的方式可增加对参数存在性的校验
    复杂性的json数据格式请直接使用官方json解析结构体，这种情况下，对参数的校验也要求不高
    经测验解析json比form-data性能要高5%左右

## protoc文件处理 集成到 Makefile文件中

    仅生成调用结构
    protoc --go_out=. ServerLinkPacket.proto

    生成客户端 服务端 及调用结构
    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative hello.proto

## 根据信号对程序进行 性能采样 [弃用]
    根据信号对程序进行 打印堆栈信息
    在默认支持http_pprof的情况下，该方案没啥优势

## 重构入口层，独立出appServer用来处理信号和资源回收 完成

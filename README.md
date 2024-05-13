# swagger
go-web框架

## 待完善计划表

## 实现路由上保存调用次数：参考这个库https://github.com/alphadose/haxmap实现
    该map仅适用于极少并发写入，大量并发修改

    在router的定义中 增加一个字段 atomic.Int64

## 定时处理器
    处理器提供rpc和http服务 可用来管理定时任务，查看日志等
        自定义时间轮
        需要rpc双向流来处理

            任务 需要支持 
                1. rpc调用 
                2. http调用
                3. 双向rpc数据流 方法绑定通知

    引申出来 当前框架需同时支持rpc和http
        能否自动根据路由 生成对应的pb文件
        
    服务重启 这样子就需要根据配置 确定有多少个网络端口需要处理重启了
        

    类似于redis nats的服务，app于定时服务之间rpc连接
        app 注册任务
        处理器发送指令触发任务
        

## go标准库rpc性能大概有grpc的2倍

## 路由处理中的正则匹配 是否考虑 在当前基础上 支持下正则路由匹配

    或者独立的两套规则，可供使用者自己决定

    正则匹配路由的性能要差于静态路由

## 对外开放的方法及路由规则 调整为显式加载 减少隐式加载的黑洞

## http pprof内置支持
    
    可自定义重写路由规则 方便追加验证规则 加强安全性

## /debug/var内置支持

    支持自定义路由规则

## 版本信息 是否也能默认加载 减少隐式加载的黑洞

## 搞个0配置启动，内置缓存处理器
    当没有redis时，使用内置的引擎，内存缓存
        性能上不如
        不支持持久化
            需要过期支持

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

## 路由处理 - 改为链表 匹配性能提升100倍 整体提升没多大[100µs] 完成
    将正则规则转换为字符串处理
    服务启动时 处理完所有路由规则 并将其注册到服务中
    利用链表存储路由规则

    正则匹配下的耗时36.827µs
    链表匹配下的耗时500ns

    handler 先生成默认的路由映射规则
        handler 可追加内置处理 比如pprof
    router 再对规则二次定义
        初始化后对handler做删除，使handler全部移交给router处理

    这次改版有点蛋疼：mac mini 六核intel core i7 32G内存
        同样请求 /welcome/index
        同样返回 json: hello world
        同样终端打印日志
            gin qps 3800
            swagger regexp qps 3550
            swagger mapLink qps 3840

        关闭日志：
            gin qps 4100 [仅关了终端彩色日志，没找到彻底关闭的方法]
            swagger regexp qps  4050 4165
            swagger mapLink qps 4200

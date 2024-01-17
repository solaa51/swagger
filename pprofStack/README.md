## 安装graphviz

    只是为了便捷查看，没有也不影响
    brew install graphviz

## 根据信号对程序进行 性能采样

    接收信号量31 
        开始采样：kill -31 pid
        结束采样：kill -31 pid

    获取到三个文件：cpu.pprof mem.pprof runtime.trace

    开始分析：
        方式一：
            go tool pprof cpu.pprof  // go tool pprof memory.pprof
            top5 指令可查看cpu占用最多的5个指令
            web 指令会调用graphviz生成svg图片并自动打开

            go tool trace runtime.trace

        方式二：
            go tool pprof -http localhost:3000 cpu.pprof
            go tool pprof -http localhost:3000 memory.pprof

## 根据信号对程序进行 打印堆栈信息

    接收信号量32
        开始采样：kill -32 pid
        结束采样：kill -32 pid

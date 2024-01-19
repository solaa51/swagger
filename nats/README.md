nats 消息队列处理支持分布式使用
    ## RequestReplayConsumer
        普通 请求-回应模式 订阅主题
        有消息处理时重置等待时间，无消息处理后到期关闭消费者 [不确定性太大 不推荐使用]
        如果多开消费者，会都收到消息 [需考虑重复处理]

## 普通模式下的消息发送和处理案例 [同步发送并接收返回]
    
        nt := nats.NewNats()
    
        // 订阅主题
        err := nt.RequestReplayConsumer("mugame_serverlink_test.>", func(subject string, body []byte) []byte {
            return []byte("我收到你的消息了:" + string(body))
        })
        if err != nil {
            return nil, err
        }
        
        //发送消息
        subject := fmt.Sprintf("mugame_serverlink_%s.group.group_%d.ProcessSignal", appEnv, regionId)
        data := []byte(fmt.Sprintf("%d|%d|%d|0|0", lineId, signal, data_1))
    
        ret, err := nt.RequestReplayRequest(subject, data)

        //关闭连接 - 跟消费者一起启动时，不要调用关闭
        nc.Close()

## 流处理使用案例

    func main() {
        nt := nats.NewNats()
    
        for i := 1; i < 3; i++ {
            go func(i int) {
                err := nt.StreamConsumer("greetStream", "greet.>", "greetStreamConsumer", func(subject string, body []byte) {
                    fmt.Println("处理器", i, subject, string(body))
                })
                if err != nil {
                    panic(err)
                }
            }(i)
        }
    
        go func() {
            for i := 0; i < 10; i++ {
                go func(i int) {
                    time.Sleep(time.Second * time.Duration(cFunc.RandRangeInt(1, 10)))
                    err := nt.StreamPublish("greet.testSend", []byte("消息："+strconv.Itoa(i)+" "+cFunc.Date("Y-m-d H:i:s", 0)))
                    if err != nil {
                        fmt.Println("消息发送失败", err)
                    }
                }(i)
            }
        }()
    
        select {}
    }

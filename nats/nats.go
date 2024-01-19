package nats

import (
	"context"
	"errors"
	ants2 "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/solaa51/swagger/appPath"
	"github.com/solaa51/swagger/log/bufWriter"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

// nats 调用 发送消息 和 消费者

type Nats struct {
	nc *ants2.Conn
}

// Config nats配置文件结构
type Config struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func NewNats() (*Nats, error) {
	f, err := os.ReadFile(appPath.ConfigDir() + "nats.yaml")
	if err != nil {
		bufWriter.Fatal("无法获取配置文件", err.Error())
		return nil, err
	}

	natsConfig := &Config{}

	err = yaml.Unmarshal(f, natsConfig)
	if err != nil {
		bufWriter.Error("无法解析nats连接信息", err.Error())
		return nil, err
	}

	//按配置 连接ants-server
	nc, err := ants2.Connect("nats://" + natsConfig.Host + ":" + natsConfig.Port)
	if err != nil {
		bufWriter.Error("NATS服务连接失败：", err)
		return nil, err
	}

	return &Nats{
		nc: nc,
	}, nil
}

// AnswerConsumer 普通 请求-回应模式 同步消费者
// 注意： 多开所有消费者都会收到消息，注意事务处理
func (n *Nats) AnswerConsumer(subject string, fn func(subject string, body []byte) []byte) error {
	ch := make(chan error)

	go func() {
		_, err := n.nc.Subscribe(subject, func(msg *ants2.Msg) {
			reply := fn(msg.Subject, msg.Data)
			_ = msg.Respond(reply)
		})
		if err != nil {
			ch <- err
			return
		}

		//通知外层服务已开启
		close(ch)
	}()

	err, ok := <-ch
	if !ok {
		return nil
	}

	return err
}

// RequestReplayConsumer 普通 请求-回应模式 同步消费者 多开所有消费者都会收到消息
// timeoutSecond 定义等待时间，有消息处理时重置时间，无消息处理后到期关闭消费者
// 仅仅适用与消息非常少的情况 - 不推荐适用
/*func (n *Nats) RequestReplayConsumer(subject string, fn func(subject string, body []byte) []byte) error {
	ch := make(chan error)

	go func() {
		timeoutSecond := 30
		t := time.NewTimer(time.Duration(timeoutSecond) * time.Second)

		sub, err := n.nc.Subscribe(subject, func(msg *ants2.Msg) {
			t.Reset(time.Duration(timeoutSecond) * time.Second)
			reply := fn(msg.Subject, msg.Data)
			_ = msg.Respond(reply)
		})
		if err != nil {
			ch <- err
			return
		}

		//通知外层服务已开启
		close(ch)

		for {
			select {
			case <-t.C:
				_ = sub.Unsubscribe()
				n.Close()
			}
		}
	}()

	err, ok := <-ch
	if !ok {
		return nil
	}

	return err
}*/

// RequestReplayRequest 普通 请求-回应模式 发送请求
func (n *Nats) RequestReplayRequest(subject string, body []byte) ([]byte, error) {
	rep, err := n.nc.Request(subject, body, time.Second*5)
	if err != nil {
		return nil, err
	}

	return rep.Data, err
}

// StreamConsumer 流模式 按主题过滤并持久化消息 消费者
// subject 主题订阅规则
// durable 持久消费者：此参数为空时，相当与广播,所有消费者都可以接收到；不为空时，消息只会被一个消费者处理
// 仅仅适用与消息非常少的情况 - 不推荐适用
/*func (n *Nats) StreamConsumer(streamName string, subject string, durable string, fn func(subject string, body []byte)) error {
	js, err := jetstream.New(n.nc)
	if err != nil {
		return err
	}

	//声明流配置 游戏线路配置
	cfg := jetstream.StreamConfig{
		Name:      streamName,
		Retention: jetstream.InterestPolicy, //跟配置无关的消息将被丢弃
		Subjects:  []string{subject},        //一个流可以绑定多个主题, 此处只绑定一个主题
	}
	cfg.Storage = jetstream.FileStorage //配置持久化方式为 文件存储

	//超时和取消context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//创建流
	stream, err := js.CreateStream(ctx, cfg)
	if err != nil {
		return err
	}

	cons, _ := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:   durable,
		AckPolicy: jetstream.AckExplicitPolicy,
	})

	ch := make(chan error)

	go func() {
		timeoutSecond := 30
		t := time.NewTimer(time.Duration(timeoutSecond) * time.Second)

		cc, err := cons.Consume(func(msg jetstream.Msg) {
			_ = msg.Ack()
			fn(msg.Subject(), msg.Data())
		})

		if err != nil {
			ch <- err
			return
		}

		//通知外层服务已开启
		close(ch)

		for {
			select {
			case <-t.C:
				cc.Stop()
				n.Close()
			}
		}
	}()

	err, ok := <-ch
	if !ok {
		return nil
	}

	return err
}*/

// StreamPublish 流模式 发送消息
func (n *Nats) StreamPublish(subject string, body []byte) error {
	js, err := jetstream.New(n.nc)
	if err != nil {
		return errors.New("nats-jetstream创建失败：" + err.Error())
	}

	//超时和取消context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = js.Publish(ctx, subject, body)
	if err != nil {
		return errors.New("nats publish失败：" + err.Error())
	}

	return nil
}

func (n *Nats) Close() {
	_ = n.nc.Drain()
}

package natsv2

import (
	"context"
	"errors"
	ants2 "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/solaa51/swagger/app"
	"github.com/solaa51/swagger/appPath"
	"github.com/solaa51/swagger/log/bufWriter"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

// nats 调用 发送消息 和 消费者
var defaultNats *Nats
var Conf *Config

func AnswerConsumer(subject string, fn func(subject string, body []byte) []byte) error {
	return defaultNats.AnswerConsumer(subject, fn)
}

func RequestReplayRequest(subject string, body []byte) ([]byte, error) {
	return defaultNats.RequestReplayRequest(subject, body)
}

func StreamPublish(subject string, body []byte) error {
	return defaultNats.StreamPublish(subject, body)
}

type Nats struct {
	nc *ants2.Conn
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

// RequestReplayRequest 普通 请求-回应模式 发送请求
func (n *Nats) RequestReplayRequest(subject string, body []byte) ([]byte, error) {
	rep, err := n.nc.Request(subject, body, time.Second*5)
	if err != nil {
		return nil, err
	}

	return rep.Data, err
}

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

// Config nats配置文件结构
type Config struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func newConfig() (*Config, error) {
	f, err := os.ReadFile(appPath.ConfigDir() + "nats.yaml")
	if err != nil {
		return nil, err
	}

	natsConfig := &Config{}
	err = yaml.Unmarshal(f, natsConfig)
	if err != nil {
		bufWriter.Error("无法解析nats连接信息", err.Error())
		return nil, err
	}

	return natsConfig, nil
}

func NewClient(host, port string) (*Nats, error) {
	//按配置 连接ants-server
	nc, err := ants2.Connect("nats://" + host + ":" + port)
	if err != nil {
		bufWriter.Error("NATS服务连接失败：", err)
		return nil, err
	}

	return &Nats{
		nc: nc,
	}, nil
}

func init() {
	Conf, err := newConfig()
	if err != nil {
		bufWriter.Fatal("无法解析配置文件", err.Error())
	}

	defaultNats, err = NewClient(Conf.Host, Conf.Port)
	if err != nil {
		bufWriter.Fatal(err.Error())
	}

	app.RegistClose(defaultNats.Close)
}

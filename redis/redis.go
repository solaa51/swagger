package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/solaa51/swagger/app"
	"github.com/solaa51/swagger/appPath"
	"github.com/solaa51/swagger/log/bufWriter"
	"github.com/solaa51/swagger/watchConfig"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
	"time"
)

type Client struct {
	*redis.Client
}

// Get bool表示是否存在该key
func (c *Client) Get(key string) (string, bool, error) {
	val, err := c.Client.Get(context.Background(), key).Result()

	if errors.Is(err, redis.Nil) {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}

	return val, true, nil
}

// Set 0-second表示为没有过期时间
func (c *Client) Set(key string, value string, second int) error {
	return c.Client.Set(context.Background(), key, value, time.Second*time.Duration(second)).Err()
}

func (c *Client) IsExists(key string) bool {
	i, err := c.Client.Exists(context.Background(), key).Result()
	if err != nil {
		return false
	}

	if i == 1 {
		return true
	} else {
		return false
	}
}

// HSet hash表 any可为slice[成对] map等类型 当second=0时无过期时间
func (c *Client) HSet(key string, value any, second int) error {
	err := c.Client.HSet(context.Background(), key, value).Err()
	if err != nil {
		return err
	}

	if second > 0 {
		c.Client.Expire(context.Background(), key, time.Second*time.Duration(second))
	}

	return nil
}

// HSetNX hash表增加一个键值对
func (c *Client) HSetNX(key string, field string, value any) error {
	return c.Client.HSetNX(context.Background(), key, field, value).Err()
}

// HIncrBy hash表字段值 增减
func (c *Client) HIncrBy(key string, field string, incr int64) (int64, error) {
	return c.Client.HIncrBy(context.Background(), key, field, incr).Result()
}

func (c *Client) HDel(key string, field string) (int64, error) {
	return c.Client.HDel(context.Background(), key, field).Result()
}

// HGetAll 获取hash列表中的所有键值对
func (c *Client) HGetAll(key string) (map[string]string, error) {
	return c.Client.HGetAll(context.Background(), key).Result()
}

func (c *Client) HGetFieldValue(key string, field string) (string, error) {
	return c.Client.HGet(context.Background(), key, field).Result()
}

// TTL 获取剩余时间
func (c *Client) TTL(key string) (float64, error) {
	val, err := c.Client.TTL(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}

	return val.Seconds(), nil
}

func (c *Client) Expire(key string, second int) {
	c.Client.Expire(context.Background(), key, time.Second*time.Duration(second))
}

func (c *Client) Del(key string) error {
	if c.Client == nil {
		return errors.New("无效的redis连接")
	}

	return c.Client.Del(context.Background(), key).Err()
}

func (c *Client) Close() {
	if c.Client == nil {
		return
	}

	_ = c.Client.Close()
}

var defaultClient *Client
var Conf *Config     //redis配置信息
var keyPrefix string //前缀
var wg sync.Mutex

// Config redis配置文件结构
type Config struct {
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	User   string `yaml:"user"`
	Pass   string `yaml:"pass"`
	DB     int    `yaml:"db"`
	Prefix string `yaml:"prefix"`
}

func newConfig() (*Config, error) {
	f, err := os.ReadFile(appPath.ConfigDir() + "redis.yaml")
	if err != nil {
		return nil, errors.New("无法读取redis配置文件")
	}

	Conf = &Config{}
	err = yaml.Unmarshal(f, Conf)
	if err != nil {
		return nil, errors.New("无法解析redis连接信息:" + err.Error())
	}

	return Conf, nil
}

func init() {
	Conf, err := newConfig()
	if err != nil {
		bufWriter.Fatal(err.Error())
	}
	keyPrefix = Conf.Prefix

	defaultClient, err = NewClient(Conf.Host, Conf.Port, Conf.User, Conf.Pass, Conf.DB)
	if err != nil {
		bufWriter.Fatal(err.Error())
	}

	ch, _ := watchConfig.AddWatch(appPath.ConfigDir() + "redis.yaml")
	go func() {
		for {
			select {
			case <-ch:
				bufWriter.Info(appPath.ConfigDir()+"redis.yaml", "文件变更触发更新")
				cc, err := newConfig()
				if err == nil {
					wg.Lock()
					upClient, err := NewClient(Conf.Host, Conf.Port, Conf.User, Conf.Pass, Conf.DB)
					if err == nil {
						defaultClient = upClient
						Conf = cc
						keyPrefix = cc.Prefix
					} else {
						bufWriter.Error("redis配置变更触发更新失败：", err.Error())
					}
					wg.Unlock()
				} else {
					bufWriter.Error("redis配置变更触发更新失败：", err.Error())
				}
			}
		}
	}()

	app.RegistClose(defaultClient.Close)
}

// KeyPrefix 前缀配置
func KeyPrefix() string {
	return keyPrefix
}

func IsExists(key string) bool {
	return defaultClient.IsExists(key)
}

func HSet(key string, value any, second int) error {
	return defaultClient.HSet(key, value, second)
}

func HSetNX(key string, field string, value any) error {
	return defaultClient.HSetNX(key, field, value)
}

func HIncrBy(key string, field string, incr int64) (int64, error) {
	return defaultClient.HIncrBy(key, field, incr)
}

func HDel(key string, field string) (int64, error) {
	return defaultClient.HDel(key, field)
}

func HGetAll(key string) (map[string]string, error) {
	return defaultClient.HGetAll(key)
}

func HGetFieldValue(key string, field string) (string, error) {
	return defaultClient.HGetFieldValue(key, field)
}

func Set(key string, value string, second int) error {
	return defaultClient.Set(key, value, second)
}

func Get(key string) (string, bool, error) {
	return defaultClient.Get(key)
}

func TTL(key string) (float64, error) {
	return defaultClient.TTL(key)
}

func Expire(key string, second int) {
	defaultClient.Expire(key, second)
}

// Del 删除
func Del(key string) error {
	return defaultClient.Del(key)
}

func Close() {
	defaultClient.Close()
}

func NewClient(host, port, name, pass string, db int) (*Client, error) {
	cc := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Username: name,
		Password: pass,
		DB:       db,
		//PoolSize: 20, //默认为 cpu核数*10
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := cc.Ping(ctx).Result()
	if err != nil {
		return nil, errors.New("连接redis失败:" + err.Error())
	}

	cC := &Client{cc}

	return cC, nil
}

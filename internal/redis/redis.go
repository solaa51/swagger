package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"os"
	"swagger/internal/appPath"
	"swagger/internal/log/bufWriter"
	"time"
)

var iClient *redis.Client

// 前缀
var keyPrefix string

// Config redis配置文件结构
type Config struct {
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	User   string `yaml:"user"`
	Pass   string `yaml:"pass"`
	DB     int    `yaml:"db"`
	Prefix string `yaml:"prefix"`
}

func init() {
	cc := &Config{}

	f, err := os.ReadFile(appPath.ConfigDir() + "redis.yaml")
	if err != nil {
		bufWriter.Fatal("无法解析redis连接信息", err.Error())
	}

	err = yaml.Unmarshal(f, cc)
	if err != nil {
		bufWriter.Fatal("无法解析redis连接信息", err.Error())
	}

	iClient, err = NewClient(cc.Host, cc.Port, cc.User, cc.Pass, cc.DB)
	if err != nil {
		bufWriter.Fatal(err.Error())
		return
	}

	keyPrefix = cc.Prefix
}

// Db 自定义调用请先使用该方法获取实例
func Db() *redis.Client {
	return iClient
}

// KeyPrefix 前缀配置
func KeyPrefix() string {
	return keyPrefix
}

func IsExists(key string) bool {
	i, err := iClient.Exists(context.Background(), key).Result()
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
func HSet(key string, value any, second int) error {
	err := iClient.HSet(context.Background(), key, value).Err()
	if err != nil {
		return err
	}

	if second > 0 {
		iClient.Expire(context.Background(), key, time.Second*time.Duration(second))
	}

	return nil
}

// HGetAll 获取hash列表中的所有键值对
func HGetAll(key string) (map[string]string, error) {
	return iClient.HGetAll(context.Background(), key).Result()
}

func HGetFieldValue(key string, field string) (string, error) {
	return iClient.HGet(context.Background(), key, field).Result()
}

// Set 0-second表示为没有过期时间
func Set(key string, value string, second int) error {
	return iClient.Set(context.Background(), key, value, time.Second*time.Duration(second)).Err()
}

// Get bool表示是否存在该key
func Get(key string) (string, bool, error) {
	val, err := iClient.Get(context.Background(), key).Result()

	if errors.Is(err, redis.Nil) {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}

	return val, true, nil
}

// TTL 获取剩余时间
func TTL(key string) (float64, error) {
	val, err := iClient.TTL(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}

	return val.Seconds(), nil
}

func Expire(key string, second int) {
	iClient.Expire(context.Background(), key, time.Second*time.Duration(second))
}

// Del 删除
func Del(key string) error {
	if iClient == nil {
		return errors.New("无效的redis连接")
	}

	return iClient.Del(context.Background(), key).Err()
}

func NewClient(host, port, name, pass string, db int) (*redis.Client, error) {
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

	return cc, nil
}

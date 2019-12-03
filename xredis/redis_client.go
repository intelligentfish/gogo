package xredis

import (
	"github.com/go-redis/redis"
	"github.com/intelligentfish/gogo/app_cfg"
	"sync"
)

// 变量
var (
	clientOnce     sync.Once // 执行一次
	clientInstance *Client   // 单实例
)

// RedisClient Redis客户端
type Client struct {
	C *redis.Client
}

// newRedisClient 工厂方法
func newRedisClient() *Client {
	return &Client{}
}

// InitializeWithTimeout 初始化方法
func (object *Client) Initialize() (err error) {
	appCfg := app_cfg.GetInstance()
	object.C = redis.NewClient(&redis.Options{
		Addr:     appCfg.RedisCfg.Address,
		Password: appCfg.RedisCfg.Password,
		DB:       appCfg.RedisCfg.DB,
	})
	_, err = object.C.Ping().Result()
	return
}

// GetRedisClientInstance 单实例
func GetRedisClientInstance() *Client {
	clientOnce.Do(func() {
		clientInstance = newRedisClient()
	})
	return clientInstance
}

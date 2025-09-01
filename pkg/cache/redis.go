package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Instance struct {
	MasterClient redis.UniversalClient
}

type RedisConfig struct {
	Addr        string
	Password    string
	DB          int
	ClusterMode bool
	PoolSize    int
}

var mainIns *Instance

// 从环境变量初始化Redis连接
func Init() {
	config := GetRedisConfigFromEnv()
	mainIns = createDataSource(config)

}

// 获取主Redis客户端
func MainRedis() redis.UniversalClient {
	return mainIns.MasterClient
}

// 创建数据源
func createDataSource(config RedisConfig) *Instance {
	dataSource := &Instance{}

	if config.ClusterMode {
		// 集群模式
		clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:           []string{config.Addr},
			DialTimeout:     time.Second * 10,
			ConnMaxIdleTime: time.Second * 5,
			PoolSize:        config.PoolSize,
		})

		status, err := clusterClient.Ping(context.Background()).Result()
		if status != "PONG" {
			msg := fmt.Sprintf("can not ping redis. config: %+v. status: %v. err: %v", config, status, err)
			panic(msg)
		}

		dataSource.MasterClient = clusterClient
	} else {
		// 单机模式
		cfg := &redis.Options{}
		cfg.Addr = config.Addr
		cfg.Password = config.Password
		cfg.DB = config.DB
		cfg.DialTimeout = time.Second * 10
		cfg.ConnMaxIdleTime = time.Second * 5
		if config.PoolSize != 0 {
			cfg.PoolSize = config.PoolSize
		}

		client := redis.NewClient(cfg)
		status := client.Ping(context.Background()).Val()
		if status != "PONG" {
			msg := fmt.Sprintf("can not ping redis. config: %+v. status: %v", config, status)
			panic(msg)
		}
		dataSource.MasterClient = client
	}

	return dataSource
}

// 便捷方法：设置缓存
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return MainRedis().Set(ctx, key, value, expiration).Err()
}

// 便捷方法：获取缓存
func Get(ctx context.Context, key string) (string, error) {
	return MainRedis().Get(ctx, key).Result()
}

// 便捷方法：删除缓存
func Del(ctx context.Context, keys ...string) error {
	return MainRedis().Del(ctx, keys...).Err()
}

// 便捷方法：检查键是否存在
func Exists(ctx context.Context, keys ...string) (int64, error) {
	return MainRedis().Exists(ctx, keys...).Result()
}

// 便捷方法：设置过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return MainRedis().Expire(ctx, key, expiration).Err()
}

// 便捷方法：获取剩余过期时间
func TTL(ctx context.Context, key string) (time.Duration, error) {
	return MainRedis().TTL(ctx, key).Result()
}

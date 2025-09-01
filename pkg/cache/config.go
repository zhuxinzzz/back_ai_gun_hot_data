package cache

import (
	"os"
	"strconv"
)

// 从环境变量获取Redis配置
func GetRedisConfigFromEnv() RedisConfig {
	//return RedisConfig{
	//	Addr:        getEnv("REDIS_ADDR", "hsy:16379"),
	//	Password:    getEnv("REDIS_PASSWORD", "redisInstance1!"),
	//	DB:          getEnvAsInt("REDIS_DB", 0),
	//	ClusterMode: getEnvAsBool("REDIS_CLUSTER_MODE", false),
	//	PoolSize:    getEnvAsInt("REDIS_POOL_SIZE", 10),
	//}
	return RedisConfig{
		Addr: getEnv("REDIS_ADDR", "satoshi8.redis.rds.aliyuncs.com:6379"),
		//Password:    getEnv("REDIS_PASSWORD", "redisInstance1!"),
		DB:          getEnvAsInt("REDIS_DB", 15), // 15 is test db.
		ClusterMode: getEnvAsBool("REDIS_CLUSTER_MODE", false),
		PoolSize:    getEnvAsInt("REDIS_POOL_SIZE", 10),
	}
}

// 工具函数：获取环境变量，提供默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 工具函数：获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 工具函数：获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

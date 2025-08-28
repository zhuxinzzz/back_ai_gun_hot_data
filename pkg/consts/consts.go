package consts

import (
	"os"
	"strconv"
	"time"
)

// 从环境变量读取配置的函数
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

var (
	PG_HOST     = getEnvOrDefault("PG_HOST", "")
	PG_PORT     = getEnvIntOrDefault("PG_PORT", 0)
	PG_NAME     = getEnvOrDefault("PG_NAME", "")
	PG_USER     = getEnvOrDefault("PG_USER", "")
	PG_PASSWORD = getEnvOrDefault("PG_PASSWORD", "")
	PG_SSLMODE  = getEnvOrDefault("PG_SSLMODE", "")
)

// 批处理配置 - 从环境变量读取
var (
	BATCH_SIZE = getEnvIntOrDefault("BATCH_SIZE", 1)
	API_DELAY  = time.Duration(getEnvIntOrDefault("API_DELAY", 5200)) * time.Millisecond // 毫秒
)

// API来源常量
const (
	SOURCE_API_COINGECKO = "coin_gecko"
	SOURCE_API_BINANCE   = "binance"
	SOURCE_API_CMC       = "coin_market_cap"
)

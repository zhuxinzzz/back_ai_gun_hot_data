package dao

import "github.com/redis/go-redis/v9"

var (
	mainResit *redis.UniversalClient
)

func GetMainRedis() *redis.UniversalClient {
	return mainResit
}

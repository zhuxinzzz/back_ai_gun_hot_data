package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"back_ai_gun_data/consumer"
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/lr"
)

func main() {
	lr.Init()

	redisConfig := cache.GetRedisConfigFromEnv()
	cache.Init(redisConfig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.StartAllConsumers(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	cancel()
}

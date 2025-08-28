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
	// 初始化logger
	lr.Init()

	// 初始化缓存
	redisConfig := cache.GetRedisConfigFromEnv()
	cache.Init(redisConfig)

	config := consumer.GetConfigFromEnv()

	c, err := consumer.NewConsumer(config)
	if err != nil {
		lr.E().Fatalf("Failed to create consumer: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := c.Start(ctx); err != nil {
			lr.E().Errorf("Consumer error: %v", err)
		}
	}()

	<-sigChan
	cancel()
}

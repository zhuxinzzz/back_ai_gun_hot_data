package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"back_ai_gun_data/consumer"
)

func main() {
	config := consumer.GetConfigFromEnv()

	c, err := consumer.NewConsumer(config)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := c.Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	//log.Println("Consumer service started. Press Ctrl+C to stop.")

	<-sigChan
	//log.Println("Shutting down consumer service...")
	cancel()
}

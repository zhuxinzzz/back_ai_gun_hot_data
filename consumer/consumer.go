package consumer

import (
	"back_ai_gun_data/pkg/model"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartMainConsumer(ctx context.Context) {
	go func() {
		if err := startConsumer(ctx); err != nil {
			lr.E().Errorf("Consumer error: %v", err)
		}
	}()
}

func startConsumer(ctx context.Context) error {
	conn, err := amqp.Dial(getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"))
	if err != nil {
		lr.E().Errorf("Failed to connect to RabbitMQ: %v", err)
		return fmt.Errorf("connect failed: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		lr.E().Errorf("Failed to open channel: %v", err)
		return fmt.Errorf("channel failed: %w", err)
	}
	defer ch.Close()

	err = ch.Qos(getEnvInt("PREFETCH", 10), 0, false)
	if err != nil {
		lr.E().Errorf("Failed to set QoS: %v", err)
		return fmt.Errorf("qos failed: %w", err)
	}

	queueName := getEnv("QUEUE_NAME", "etl-entity-data")
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		lr.E().Errorf("Failed to declare queue: %v", err)
		return fmt.Errorf("queue declare failed: %w", err)
	}

	// 处理历史消息
	if err := processHistoricalMessages(ch, queueName); err != nil {
		lr.E().Errorf("Failed to process historical messages: %v", err)
		// 不返回错误，继续处理新消息
	}

	consumerTag := fmt.Sprintf("etl-consumer-token-data-%d", os.Getpid())
	msgs, err := ch.Consume(queueName, consumerTag, false, false, false, false, nil)
	if err != nil {
		lr.E().Errorf("Failed to register consumer: %v", err)
		return fmt.Errorf("consume failed: %w", err)
	}

	lr.I().Infof("Consumer started, listening on queue: %s", queueName)

	var semaphore = make(chan struct{}, getEnvInt("MAX_CONCURRENT", 100))

	for {
		select {
		case <-ctx.Done():
			lr.I().Info("Consumer context cancelled, stopping...")
			return nil
		case msg := <-msgs:
			semaphore <- struct{}{} // 获取信号量
			go func(msg amqp.Delivery) {
				defer func() { <-semaphore }() // 释放信号量
				handleMsg(msg)
			}(msg)
		}
	}
}

func handleMsg(msg amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			lr.E().Errorf("Panic in message handler: %v", r)
			err := msg.Nack(false, true)
			if err != nil {
				lr.E().Error(err)
				return
			} // 重新入队
		}
	}()

	var messageData model.MessageData
	if err := json.Unmarshal(msg.Body, &messageData); err != nil {
		lr.E().Errorf("Failed to unmarshal message: %v", err)
		err := msg.Nack(false, false)
		if err != nil {
			lr.E().Error(err)
			return
		} // 拒绝消息，不重新入队
		return
	}

	if err := services.ProcessMessageData(&messageData); err != nil {
		lr.E().Errorf("Failed to process message: %v", err)
		err := msg.Nack(false, true)
		if err != nil {
			lr.E().Error(err)
			return
		} // 重新入队
		return
	}

	err := msg.Ack(false)
	if err != nil {
		lr.E().Error(err)
		return
	}
}

func processHistoricalMessages(ch *amqp.Channel, queueName string) error {
	lr.I().Infof("Processing historical messages from queue: %s", queueName)

	processedCount := 0
	maxHistoricalMessages := getEnvInt("MAX_HISTORICAL_MESSAGES", 1000) // 限制处理的历史消息数量

	for processedCount < maxHistoricalMessages {
		// 使用Get方法获取消息，不自动确认
		msg, ok, err := ch.Get(queueName, false)
		if err != nil {
			return fmt.Errorf("failed to get message: %w", err)
		}

		// 如果没有更多消息，退出循环
		if !ok {
			break
		}

		// 处理消息
		handleMsg(msg)
		processedCount++

		// 每处理100条消息输出一次进度
		if processedCount%100 == 0 {
			lr.I().Infof("Processed %d historical messages", processedCount)
		}
	}

	if processedCount > 0 {
		lr.I().Infof("Finished processing %d historical messages", processedCount)
	} else {
		lr.I().Info("No historical messages found")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

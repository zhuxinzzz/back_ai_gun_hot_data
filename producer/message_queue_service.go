package producer

import (
	"back_ai_gun_data/pkg/consts"
	"back_ai_gun_data/pkg/lr"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	conn    *amqp.Connection
	channel *amqp.Channel
	mutex   sync.RWMutex
)

// Connect 连接到RabbitMQ
func Connect() error {
	mutex.Lock()
	defer mutex.Unlock()

	if conn != nil && !conn.IsClosed() {
		return nil
	}

	var err error
	conn, err = amqp.Dial(consts.DEFAULT_RABBITMQ_URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err = conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明队列
	_, err = channel.QueueDeclare(
		consts.QUEUE_AI_TOKEN, // queue name
		true,                  // durable
		false,                 // auto-deleted
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	lr.I().Info("Successfully connected to RabbitMQ and declared ai_token_queue")
	return nil
}

// Close 关闭连接
func Close() error {
	mutex.Lock()
	defer mutex.Unlock()

	if channel != nil {
		if err := channel.Close(); err != nil {
			lr.E().Errorf("Failed to close channel: %v", err)
		}
		channel = nil
	}

	if conn != nil {
		if err := conn.Close(); err != nil {
			lr.E().Errorf("Failed to close connection: %v", err)
		}
		conn = nil
	}

	return nil
}

// ensureConnection 确保连接可用
func ensureConnection() error {
	if conn == nil || conn.IsClosed() || channel == nil || channel.IsClosed() {
		return Connect()
	}
	return nil
}

func SendNewTokensMessage(ctx context.Context, newTokens interface{}) error {
	if err := ensureConnection(); err != nil {
		return fmt.Errorf("failed to ensure connection: %w", err)
	}

	// 构造消息体
	messageBody := map[string]interface{}{
		"entities": newTokens,
	}

	// 序列化消息
	body, err := json.Marshal(messageBody)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发送消息
	err = channel.PublishWithContext(ctx,
		"",                    // exchange
		consts.QUEUE_AI_TOKEN, // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 消息持久化
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	lr.I().Infof("Successfully sent new tokens message to %s, tokens count: %v", consts.QUEUE_AI_TOKEN, newTokens)
	return nil
}

func SendNewTokensMessageAsync(ctx context.Context, newTokens interface{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				lr.E().Errorf("Panic in SendNewTokensMessageAsync: %v", r)
			}
		}()

		if err := SendNewTokensMessage(ctx, newTokens); err != nil {
			lr.E().Errorf("Failed to send new tokens message asynchronously: %v", err)
		}
	}()
}

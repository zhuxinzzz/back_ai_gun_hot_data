package producer

import (
	"back_ai_gun_data/pkg/consts"
	"back_ai_gun_data/pkg/lr"
	"os"
	"sync"

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
	conn, err = amqp.Dial(getEnv("RABBITMQ_URL", consts.DEFAULT_RABBITMQ_URL))
	if err != nil {
		lr.E().Error(err)
		return err
	}

	channel, err = conn.Channel()
	if err != nil {
		conn.Close()
		lr.E().Error(err)
		return err
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
		lr.E().Error(err)
		return err
	}

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

// getEnv 从环境变量获取配置，如果不存在则使用默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

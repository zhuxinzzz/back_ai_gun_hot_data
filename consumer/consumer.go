package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"back_ai_gun_data/model"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

// 消费者配置
type ConsumerConfig struct {
	RabbitMQURL string
	QueueName   string
	Prefetch    int
}

// 从环境变量获取配置
func GetConfigFromEnv() ConsumerConfig {
	return ConsumerConfig{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName:   getEnv("QUEUE_NAME", "etl-entity-data"),
		Prefetch:    getEnvAsInt("PREFETCH", 10),
	}
}

// 启动主消费者
func StartMainConsumer(ctx context.Context) {
	config := GetConfigFromEnv()

	c, err := NewConsumer(config)
	if err != nil {
		lr.E().Fatalf("Failed to create consumer: %v", err)
	}
	defer c.Close()

	go func() {
		if err := c.Start(ctx); err != nil {
			lr.E().Errorf("Consumer error: %v", err)
		}
	}()
}

// 消费者实例
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  ConsumerConfig
}

// 初始化消费者
func NewConsumer(config ConsumerConfig) (*Consumer, error) {
	conn, err := amqp.Dial(config.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 设置预取数量
	err = ch.Qos(config.Prefetch, 0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		config:  config,
	}, nil
}

// 启动消费者
func (c *Consumer) Start(ctx context.Context) error {
	// 确保队列存在
	if err := c.ensureQueueExists(); err != nil {
		return fmt.Errorf("failed to ensure queue exists: %w", err)
	}

	msgs, err := c.channel.Consume(
		c.config.QueueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	lr.I().Infof("Consumer started, listening on queue: %s", c.config.QueueName)

	for {
		select {
		case <-ctx.Done():
			lr.I().Info("Consumer context cancelled, stopping...")
			return nil
		case msg := <-msgs:
			c.handleMessage(msg)
		}
	}
}

// 确保队列存在
func (c *Consumer) ensureQueueExists() error {
	// 声明队列，如果不存在则创建
	_, err := c.channel.QueueDeclare(
		c.config.QueueName, // name
		true,               // durable (持久化)
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	lr.I().Infof("Queue '%s' is ready", c.config.QueueName)
	return nil
}

// 处理消息
func (c *Consumer) handleMessage(msg amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			lr.E().Errorf("Panic in message handler: %v", r)
			msg.Nack(false, true) // 重新入队
		}
	}()

	lr.I().Infof("Received message: %s", msg.MessageId)

	var messageData model.MessageData
	if err := json.Unmarshal(msg.Body, &messageData); err != nil {
		lr.E().Errorf("Failed to unmarshal message: %v", err)
		msg.Nack(false, false) // 拒绝消息，不重新入队
		return
	}

	// 调用独立的业务处理函数
	if err := service.ProcessMessageData(&messageData); err != nil {
		lr.E().Errorf("Failed to process message: %v", err)
		msg.Nack(false, true) // 重新入队
		return
	}

	// 确认消息
	msg.Ack(false)
	lr.I().Infof("Message processed successfully: %s", messageData.ID)
}

// 关闭消费者
func (c *Consumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			lr.E().Errorf("Failed to close channel: %v", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			lr.E().Errorf("Failed to close connection: %v", err)
		}
	}

	return nil
}

// 健康检查
func (c *Consumer) HealthCheck() error {
	if c.conn == nil || c.conn.IsClosed() {
		return fmt.Errorf("connection is closed")
	}

	if c.channel == nil || c.channel.IsClosed() {
		return fmt.Errorf("channel is closed")
	}

	return nil
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
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

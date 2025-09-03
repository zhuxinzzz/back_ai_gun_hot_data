package consumer

import (
	"back_ai_gun_data/pkg/model"
	"back_ai_gun_data/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"back_ai_gun_data/pkg/consts"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartIntelligenceConsumer(ctx context.Context) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				lr.E().Errorf("Panic in intelligence consumer: %v", r)
			}
		}()

		if err := startConsumer(ctx, consts.QUEUE_INTELLIGENCE_SORT, consts.CONSUMER_TAG_INTELLIGENCE); err != nil {
			lr.E().Errorf("Intelligence consumer error: %v", err)
		}
	}()
}

// 启动ETL实体数据consumer
func StartETLEntityConsumer(ctx context.Context) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				lr.E().Errorf("Panic in ETL entity consumer: %v", r)
			}
		}()
		if err := startConsumer(ctx, consts.QUEUE_ETL_ENTITY_DATA, consts.CONSUMER_TAG_ETL_ENTITY); err != nil {
			lr.E().Errorf("ETL entity consumer error: %v", err)
		}
	}()
}

// 通用的consumer启动函数
func startConsumer(ctx context.Context, queueName, consumerTag string) error {
	conn, err := amqp.Dial(getEnv("RABBITMQ_URL", consts.DEFAULT_RABBITMQ_URL))
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

	err = ch.Qos(getEnvInt("PREFETCH", consts.DEFAULT_PREFETCH_COUNT), 0, false)
	if err != nil {
		lr.E().Errorf("Failed to set QoS: %v", err)
		return fmt.Errorf("qos failed: %w", err)
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		lr.E().Errorf("Failed to declare queue: %v", err)
		return fmt.Errorf("queue declare failed: %w", err)
	}

	// 处理历史消息
	//if err := processHistoricalMessages(ch, queueName); err != nil {
	//	lr.E().Errorf("Failed to process historical messages: %v", err)
	//	// 不返回错误，继续处理新消息
	//}

	fullConsumerTag := fmt.Sprintf("%s-%d", consumerTag, os.Getpid())
	msgs, err := ch.Consume(queueName, fullConsumerTag, false, false, false, false, nil)
	if err != nil {
		lr.E().Errorf("Failed to register consumer: %v", err)
		return fmt.Errorf("consume failed: %w", err)
	}

	//lr.I().Infof("Consumer started, listening on queue: %s with tag: %s", queueName, fullConsumerTag)

	var semaphore = make(chan struct{}, getEnvInt("MAX_CONCURRENT", consts.DEFAULT_MAX_CONCURRENT))

	for {
		select {
		case <-ctx.Done():
			lr.I().Info("Consumer context cancelled, stopping...")
			return nil
		case msg := <-msgs:
			semaphore <- struct{}{} // 获取信号量
			go func(msg amqp.Delivery) {
				defer func() {
					if r := recover(); r != nil {
						lr.E().Errorf("Panic in message handler for queue %s: %v", queueName, r)
					}
				}()
				defer func() { <-semaphore }() // 释放信号量
				handleMsg(msg, queueName)
			}(msg)
		}
	}
}

func handleMsg(msg amqp.Delivery, queueName string) {
	defer func() {
		if r := recover(); r != nil {
			lr.E().WithFields(lr.F{
				"backtrace": utils.GetStack(),
			}).Errorf("Panic in message handler for queue %s: %v", queueName, r)
			err := msg.Nack(false, true)
			if err != nil {
				lr.E().Error(err)
				return
			} // 重新入队
		}
	}()

	var err error
	switch queueName {
	case consts.QUEUE_INTELLIGENCE_SORT:
		var messageData model.IntelligenceMessage
		if err = json.Unmarshal(msg.Body, &messageData); err != nil {
			lr.E().Errorf("Failed to unmarshal intelligence message: %v", err)
			msg.Nack(false, false)
			return
		}
		err = processIntelligenceMessage(context.Background(), &messageData)
	case consts.QUEUE_ETL_ENTITY_DATA:
		var messageData model.ETLEntityMessage
		if err = json.Unmarshal(msg.Body, &messageData); err != nil {
			lr.E().Errorf("Failed to unmarshal ETL entity message: %v", err)
			msg.Nack(false, false)
			return
		}
		err = processETLEntityMessage(context.Background(), &messageData)
	default:
		lr.E().Errorf("Unknown queue type: %s", queueName)
		msg.Nack(false, false)
		return
	}

	if err != nil {
		lr.E().Errorf("Failed to process message from queue %s: %v", queueName, err)
		err := msg.Nack(false, true)
		if err != nil {
			lr.E().Error(err)
			return
		} // 重新入队
		return
	}

	err = msg.Ack(false)
	if err != nil {
		lr.E().Error(err)
		return
	}
}

func processIntelligenceMessage(ctx context.Context, messageData *model.IntelligenceMessage) error {
	return services.ProcessIntelligenceData(ctx, messageData)
}

func processETLEntityMessage(ctx context.Context, messageData *model.ETLEntityMessage) error {
	return services.ProcessETLEntityData(ctx, messageData)
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

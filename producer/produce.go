package producer

import (
	"back_ai_gun_data/pkg/lr"
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SendNewTokensMessage(ctx context.Context, newTokens interface{}) error {
	if err := ensureConnection(); err != nil {
		lr.E().Error(err)
		return err
	}

	// 构造消息体
	messageBody := map[string]interface{}{
		"entities": newTokens,
	}

	// 序列化消息
	body, err := json.Marshal(messageBody)
	if err != nil {
		lr.E().Error(err)
		return err
	}

	// 发送消息
	err = channel.PublishWithContext(ctx,
		"",               // exchange
		"ai_token_queue", // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 消息持久化
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		lr.E().Error(err)
		return err
	}

	lr.I().Infof("Successfully sent new tokens message to ai_token_queue, tokens count: %v", newTokens)
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

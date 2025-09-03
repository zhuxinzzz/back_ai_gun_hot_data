package producer

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/remote"
	"back_ai_gun_data/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// convertGmGnTokenToMessage 将 GmGnToken 转换为 NewTokensMessage
func convertGmGnTokenToMessage(token remote.GmGnToken) NewTokensMessage {
	// 生成唯一的 S3 路径，按日期分层 + UUIDv7 保证唯一性
	s3Key := fmt.Sprintf("image/%s-%s", time.Now().Format("2006-01-02"), utils.GenerateUUIDV7())

	return NewTokensMessage{
		Address:     token.Address,
		Chain:       token.Chain,
		ChainID:     token.ChainID,
		Decimals:    token.Decimals,
		Logo:        token.Logo,
		MarketCap:   token.MarketCap,
		Name:        token.Name,
		Network:     token.Network,
		PriceUSD:    token.PriceUSD,
		Symbol:      token.Symbol,
		TotalSupply: token.TotalSupply,
		Volume24h:   token.Volume24h,
		IsInternal:  token.IsInternal,
		Liquidity:   token.Liquidity,
		S3Key:       s3Key,
	}
}

func SendNewTokensMessage(ctx context.Context, newTokens []remote.GmGnToken) error {
	if err := ensureConnection(); err != nil {
		lr.E().Error(err)
		return err
	}

	// 转换为消息结构体
	var messages []NewTokensMessage
	for _, token := range newTokens {
		message := convertGmGnTokenToMessage(token)
		messages = append(messages, message)
	}

	// 构造消息体
	messageBody := map[string]interface{}{
		"entities": messages,
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
	//fmt.Println(utils.ToJson(messageBody))

	lr.I().Infof("Successfully sent new tokens message to ai_token_queue, tokens count: %v", newTokens)
	return nil
}

func SendNewTokensMessageAsync(ctx context.Context, newTokens []remote.GmGnToken) {
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

package producer

import (
	"context"
	"testing"
	"time"
)

func TestMessageQueueService(t *testing.T) {
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := Connect()
	if err != nil {
		t.Logf("Connection test skipped (RabbitMQ not available): %v", err)
		return
	}

	// 测试消息发送
	testData := []map[string]interface{}{
		{"name": "TestToken1", "address": "0x123"},
		{"name": "TestToken2", "address": "0x456"},
	}

	err = SendNewTokensMessage(ctx, testData)
	if err != nil {
		t.Errorf("Failed to send message: %v", err)
	}

	// 清理
	Close()
}

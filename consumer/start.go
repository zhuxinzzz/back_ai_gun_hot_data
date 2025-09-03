package consumer

import (
	"context"
)

// 启动所有消费者
func StartAllConsumers(ctx context.Context) {
	// 启动情报排序数据consumer
	go StartIntelligenceConsumer(ctx)

	// 启动ETL实体数据consumer
	// todo 暂时不开启
	//go StartETLEntityConsumer(ctx)
}

// 启动数据处理器消费者（示例）
func StartDataProcessConsumer(ctx context.Context) {
	// 实现数据处理的消费者逻辑
	// 可以处理不同类型的数据
}

// 启动通知消费者（示例）
func StartNotificationConsumer(ctx context.Context) {
	// 实现通知相关的消费者逻辑
	// 可以发送邮件、短信等通知
}

// 启动分析消费者（示例）
func StartAnalyticsConsumer(ctx context.Context) {
	// 实现数据分析的消费者逻辑
	// 可以处理统计、报表等任务
}

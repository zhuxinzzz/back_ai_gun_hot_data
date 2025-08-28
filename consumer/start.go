package consumer

import (
	"context"
)

// 启动所有消费者
func StartAllConsumers(ctx context.Context) {
	// 启动主消费者
	go StartMainConsumer(ctx)

	// 这里可以添加更多消费者
	// 例如：
	// go StartDataProcessConsumer(ctx)
	// go StartNotificationConsumer(ctx)
	// go StartAnalyticsConsumer(ctx)
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

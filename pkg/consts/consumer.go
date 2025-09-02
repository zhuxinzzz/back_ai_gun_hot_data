package consts

// Consumer队列配置
const (
	// 情报排序数据队列
	QUEUE_INTELLIGENCE_SORT = "dogex-sub-dev"

	// AI分析实体同步币信息队列
	QUEUE_ETL_ENTITY_DATA = "etl-entity-data"

	// RabbitMQ连接配置
	DEFAULT_RABBITMQ_URL = "amqp://consumer:biteagle8888@192.168.10.14:5672/"

	// Consumer配置
	DEFAULT_PREFETCH_COUNT          = 10
	DEFAULT_MAX_CONCURRENT          = 100
	DEFAULT_MAX_HISTORICAL_MESSAGES = 1000
)

// Consumer标签前缀
const (
	CONSUMER_TAG_INTELLIGENCE = "intelligence-consumer"
	CONSUMER_TAG_ETL_ENTITY   = "etl-entity-consumer"
)

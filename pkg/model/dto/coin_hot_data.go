package dto

import (
	"time"
)

// CoinHotData 币热数据模型
type CoinHotData struct {
	ID              string          `json:"id"`               // 币热数据ID
	EntityID        string          `json:"entity_id"`        // 实体ID
	Name            string          `json:"name"`             // 币名称
	Symbol          string          `json:"symbol"`           // 币符号
	Standard        *string         `json:"standard"`         // 代币实现标准
	Decimals        int             `json:"decimals"`         // 精度
	ContractAddress string          `json:"contract_address"` // 合约地址
	Logo            string          `json:"logo"`             // 图标URL
	Stats           CoinMarketStats `json:"stats"`            // 市场信息
	Chain           ChainInfo       `json:"chain"`            // 链信息
	IsShow          bool            `json:"is_show"`          // 是否显示（曾经进入前三）
	Ranking         int             `json:"ranking"`          // 当前排名
	HighestRanking  int             `json:"highest_ranking"`  // 历史最高排名
	FirstRankedAt   *time.Time      `json:"first_ranked_at"`  // 首次进入前三的时间
	LastRankedAt    *time.Time      `json:"last_ranked_at"`   // 最后进入前三的时间
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// CoinHotDataCache 币热数据缓存结构
type CoinHotDataCache struct {
	Coins     []CoinHotData `json:"coins"`      // 热数据币列表
	UpdatedAt time.Time     `json:"updated_at"` // 最后更新时间
}

// AdminMarketData 管理员服务市场数据结构
type AdminMarketData struct {
	CoinID           string `json:"coin_id"`            // 币ID
	Name             string `json:"name"`               // 币名称
	Symbol           string `json:"symbol"`             // 币符号
	ContractAddress  string `json:"contract_address"`   // 合约地址
	Chain            string `json:"chain"`              // 链信息
	CurrentPriceUSD  string `json:"current_price_usd"`  // 当前价格
	CurrentMarketCap string `json:"current_market_cap"` // 当前市值
	Ranking          int    `json:"ranking"`            // 排名
	IsShow           bool   `json:"is_show"`            // 是否显示
	UpdatedAt        string `json:"updated_at"`         // 更新时间
}

// AdminRankingRequest 管理员服务排序请求
type AdminRankingRequest struct {
	Coins []AdminMarketData `json:"coins"` // 需要排序的币列表
}

// AdminRankingResponse 管理员服务排序响应
type AdminRankingResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    []AdminMarketData `json:"data"` // 排序后的币列表
}

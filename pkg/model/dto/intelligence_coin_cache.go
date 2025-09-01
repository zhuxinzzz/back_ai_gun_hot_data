package dto

import (
	"fmt"
	"strings"
	"time"
)

// IntelligenceCoinCache 情报-币缓存模型
type IntelligenceCoinCache struct {
	ID              string          `json:"id"`               // project chain data id
	EntityID        string          `json:"entity_id"`        // 实体ID
	Name            string          `json:"name"`             // 币名称
	Symbol          string          `json:"symbol"`           // 币符号
	Standard        *string         `json:"standard"`         // 代币实现标准，eg erc20
	Decimals        int             `json:"decimals"`         // 精度
	ContractAddress string          `json:"contract_address"` // 合约地址
	Logo            string          `json:"logo"`             // 图标URL，转冷时到s3
	Stats           CoinMarketStats `json:"stats"`            // 市场信息
	Chain           ChainInfo       `json:"chain"`            // 链信息，不更新
	CreatedAt       CustomTime      `json:"created_at"`
	UpdatedAt       CustomTime      `json:"updated_at"`
}

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	timeStr := strings.Trim(string(b), `"`)

	formats := []string{
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err != nil {
			continue
		}

		ct.Time = t
		return nil
	}

	return fmt.Errorf("cannot parse time: %s", timeStr)
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ct.Time.Format("2006-01-02T15:04:05.000") + `"`), nil
}

// CoinMarketStats 币市场统计信息
type CoinMarketStats struct {
	WarningPriceUSD     string `json:"warning_price_usd"`     // 预警价格，不变动
	WarningMarketCap    string `json:"warning_market_cap"`    // 预警市值，不变动
	CurrentPriceUSD     string `json:"current_price_usd"`     // 当前价格，从gmgn获取
	CurrentMarketCap    string `json:"current_market_cap"`    // 当前市值，从gmgn获取
	HighestIncreaseRate string `json:"highest_increase_rate"` // 预警涨幅，历史最大值
}

// ChainInfo 链信息
type ChainInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Logo string `json:"logo"`
}

// IntelligenceCoinCacheData 情报-币缓存数据结构
type IntelligenceCoinCacheData struct {
	IntelligenceID string                  `json:"intelligence_id"`
	Coins          []IntelligenceCoinCache `json:"coins"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

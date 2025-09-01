package dto_cache

import (
	"fmt"
	"strings"
	"time"
)

// IntelligenceTokenCache 情报-币缓存模型
type IntelligenceTokenCache struct {
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
	HighestIncreaseRate string `json:"highest_increase_rate"` // 预警涨幅，历史最大值 当前市值除以预警市值
}

// ChainInfo 链信息
type ChainInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Logo string `json:"logo"`
}

// Equals 比较两个IntelligenceTokenCache是否相等
// 使用name、contract_address、chain.slug三个字段进行比较
func (c *IntelligenceTokenCache) Equals(other *IntelligenceTokenCache) bool {
	if c == nil || other == nil {
		return c == other
	}

	// 比较name（不区分大小写）
	if !strings.EqualFold(c.Name, other.Name) {
		return false
	}

	// 比较contract_address（不区分大小写，忽略0x前缀）
	cAddr := strings.ToLower(strings.TrimPrefix(c.ContractAddress, "0x"))
	otherAddr := strings.ToLower(strings.TrimPrefix(other.ContractAddress, "0x"))
	if cAddr != otherAddr {
		return false
	}

	// 比较chain.slug（不区分大小写）
	if !strings.EqualFold(c.Chain.Slug, other.Chain.Slug) {
		return false
	}

	return true
}

// GetUniqueKey 获取唯一标识符，用于map查找
// 格式：name:contract_address:chain_slug
func (c *IntelligenceTokenCache) GetUniqueKey() string {
	addr := strings.ToLower(strings.TrimPrefix(c.ContractAddress, "0x"))
	chainSlug := strings.ToLower(c.Chain.Slug)
	return fmt.Sprintf("%s:%s:%s", strings.ToLower(c.Name), addr, chainSlug)
}

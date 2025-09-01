package dto

import (
	"fmt"
	"strings"
	"time"
)

type SortRequest struct {
	IntelligenceID      string     `json:"intelligence_id"`
	IntelligenceHotData []TokenReq `json:"intelligence_hot_data"`
	TokenList           []TokenReq `json:"token_list"`
}
type TokenReq struct {
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

type IntelligenceTokenCacheResp struct {
	ID              string          `json:"id"`
	EntityID        string          `json:"entity_id"`
	Name            string          `json:"name"`
	Symbol          string          `json:"symbol"`
	Standard        *string         `json:"standard"`
	Decimals        int             `json:"decimals"`
	ContractAddress string          `json:"contract_address"`
	Logo            string          `json:"logo"`
	Stats           CoinMarketStats `json:"stats"`
	Chain           ChainInfo       `json:"chain"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

type CoinMarketStats struct {
	WarningPriceUSD     string `json:"warning_price_usd"`
	WarningMarketCap    string `json:"warning_market_cap"`
	CurrentPriceUSD     string `json:"current_price_usd"`
	CurrentMarketCap    string `json:"current_market_cap"`
	HighestIncreaseRate string `json:"highest_increase_rate"`
}

type ChainInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Logo string `json:"logo"`
}

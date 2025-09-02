package dto

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type RankReq struct {
	IntelligenceID      string        `json:"intelligence_id"`
	IntelligenceHotData []OldTokenReq `json:"intelligence_hot_data"`
	TokenList           []NewTokenReq `json:"token_list"`
}

type NewTokenReq struct {
	Address     string `json:"contractAddress"`
	Chain       string `json:"chain"`
	ChainID     int    `json:"chain_id"`
	Decimals    int    `json:"decimals"`
	Logo        string `json:"logo"`
	MarketCap   string `json:"market_cap"`
	Name        string `json:"name"`
	Network     string `json:"network"`
	PriceUSD    string `json:"price_usd"`
	Symbol      string `json:"symbol"`
	TotalSupply string `json:"total_supply"`
	Volume24h   string `json:"volume_24h"`
	IsInternal  bool   `json:"is_internal"`
	Liquidity   string `json:"liquidity"`
}

type OldTokenReq struct {
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

type IntelligenceTokenRankResp struct {
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

	// 外部API字段
	Network          string `json:"network"`
	ChainID          int    `json:"chain_id"`
	PriceUSD         string `json:"price_usd"`
	TotalSupply      string `json:"total_supply"`
	Volume24h        string `json:"volume_24h"`
	IsInternal       bool   `json:"is_internal"`
	Liquidity        string `json:"liquidity"`
	CurrentMarketCap string `json:"current_market_cap"`
}

type CoinMarketStats struct {
	WarningPriceUSD     string `json:"warning_price_usd"`
	WarningMarketCap    string `json:"warning_market_cap"`
	CurrentPriceUSD     string `json:"current_price_usd"`
	CurrentMarketCap    string `json:"current_market_cap"`
	HighestIncreaseRate string `json:"highest_increase_rate"`
}

type ChainInfo struct {
	ID        string `json:"id"`
	NetworkID string `json:"network_id"`
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Slug      string `json:"slug"`
	Logo      string `json:"logo"`
}

func (c *ChainInfo) UnmarshalJSON(data []byte) error {
	// 尝试解析为字符串
	var chainStr string
	if err := json.Unmarshal(data, &chainStr); err == nil {
		// 如果是字符串，设置到Name字段
		c.Name = chainStr
		c.Slug = chainStr
		return nil
	}

	// 如果不是字符串，尝试解析为对象
	type chainObj struct {
		ID        string `json:"id"`
		NetworkID string `json:"network_id"`
		Name      string `json:"name"`
		Symbol    string `json:"symbol"`
		Slug      string `json:"slug"`
		Logo      string `json:"logo"`
	}

	var obj chainObj
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	c.ID = obj.ID
	c.NetworkID = obj.NetworkID
	c.Name = obj.Name
	c.Symbol = obj.Symbol
	c.Slug = obj.Slug
	c.Logo = obj.Logo
	return nil
}

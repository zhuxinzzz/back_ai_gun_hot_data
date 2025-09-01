package remote

import "strings"

type GmGnToken struct {
	Name       string `json:"name"`
	Symbol     string `json:"symbol"`
	Address    string `json:"address"`
	Network    string `json:"network"`
	IsInternal bool   `json:"is_internal"`
	Logo       string `json:"logo"`
	MarketCap  string `json:"market_cap"` // 关键市场信息
	PriceUSD   string `json:"price_usd"`  // 关键市场信息
	Decimals   int    `json:"decimals"`
}

type TokenQueryResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string][]GmGnToken `json:"data"`
}

type TokenQueryParams struct {
	Q     string // 查询关键字，全称、简称、地址，多个查询用逗号分隔
	Chain string // 指定链
	Limit int    // 指定数量，默认10
	Fuzzy int    // 是否为模糊匹配，1: 是，0: 否，默认1
}

// 支持的链
var SupportedChains = map[string]struct{}{
	"solana":   {},
	"bsc":      {},
	"ethereum": {},
	"base":     {},
}

// IsSupportedChain 判断是否支持该链
func (t *GmGnToken) IsSupportedChain() bool {
	_, supported := SupportedChains[strings.ToLower(t.Network)]
	return supported
}

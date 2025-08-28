package remote

// 代币数据结构
type TokenInfo struct {
	Name       string `json:"name"`
	Symbol     string `json:"symbol"`
	Address    string `json:"address"`
	Network    string `json:"network"`
	IsInternal bool   `json:"is_internal"`
	Logo       string `json:"logo"`
	MarketCap  string `json:"market_cap"`
	PriceUSD   string `json:"price_usd"`
	Decimals   int    `json:"decimals"`
}

// API响应结构
type TokenQueryResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string][]TokenInfo `json:"data"`
}

// 查询参数结构
type TokenQueryParams struct {
	Q     string // 查询关键字，全称、简称、地址，多个查询用逗号分隔
	Chain string // 指定链
	Limit int    // 指定数量，默认10
	Fuzzy int    // 是否为模糊匹配，1: 是，0: 否，默认1
}

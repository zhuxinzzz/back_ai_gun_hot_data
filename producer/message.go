package producer

// NewTokensMessage 新代币消息结构体
type NewTokensMessage struct {
	Address     string `json:"address"`
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
	S3Key       string `json:"s3_key"`
}

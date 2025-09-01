package remote_service

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/pkg/model/dto_cache"
	"testing"

	"github.com/stretchr/testify/assert"
)

var cacheTokens = []dto_cache.IntelligenceTokenCache{
	{
		ID:       "1",
		Name:     "Bitcoin",
		Symbol:   "BTC",
		Logo:     "https://assets.coingecko.com/coins/images/1/large/bitcoin.png?1547033579",
		Decimals: 18,
		Stats: dto_cache.CoinMarketStats{
			CurrentPriceUSD:  "50000",
			CurrentMarketCap: "100000000",
			WarningPriceUSD:  "60000",
			WarningMarketCap: "120000000",
		},
	},
	{
		ID:       "2",
		Name:     "Ethereum",
		Symbol:   "ETH",
		Logo:     "https://assets.coingecko.com/coins/images/279/large/ethereum.png?1595348880",
		Decimals: 18,
		Stats: dto_cache.CoinMarketStats{
			CurrentPriceUSD:  "2000",
			CurrentMarketCap: "4000000",
			WarningPriceUSD:  "3000",
			WarningMarketCap: "5000000",
		},
	},
}

func TestCallAdminRanking(t *testing.T) {
	lr.Init()
	cache.Init()
	Init()

	intelligenceID := "0198f0a9-0e77-721b-99df-b94e851375d1"
	//rankedCoins, err := CallAdminRanking(intelligenceID, dtoCacheSliceToDTO(cacheTokens), dtoCacheSliceToDTO(cacheTokens))
	rankedCoins, err := CallAdminRanking(dto.SortRequest{
		IntelligenceID: intelligenceID,
		TokenList: []dto.TokenReq{
			{
				ID: "1",
				Chain: dto.ChainInfo{
					ID:   "1",
					Name: "Ethereum",
					Logo: "https://assets.coingecko.com/coins/images/279/large/ethereum.png?1595348880",
				},
				ContractAddress: "0xBTC",
			},
			{
				ID: "2",
				Chain: dto.ChainInfo{
					ID:   "2",
					Name: "Bitcoin",
					Logo: "https://assets.coingecko.com/coins/images/1/large/bitcoin.png?1547033579",
				},
				ContractAddress: "0xETH",
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, rankedCoins)
}

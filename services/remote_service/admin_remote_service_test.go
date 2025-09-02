package remote_service

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/pkg/model/dto_cache"
	"back_ai_gun_data/utils"
	"encoding/json"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

var cacheTokens = []dto_cache.IntelligenceToken{
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
	//rankedCoins, err := callAdminRanking(intelligenceID, dtoCacheSliceToDTO(cacheTokens), dtoCacheSliceToDTO(cacheTokens))
	rankedCoins, err := callAdminRanking(dto.RankReq{
		IntelligenceID: intelligenceID,
		TokenList: []dto.NewTokenReq{
			{
				Address: "0xBTC",
				Chain:   "ethereum",
				ChainID: 1,
				Name:    "Ethereum",
				Symbol:  "ETH",
				Network: "ethereum",
			},
			{
				Address: "0xETH",
				Chain:   "bitcoin",
				ChainID: 2,
				Name:    "Bitcoin",
				Symbol:  "BTC",
				Network: "bitcoin",
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, rankedCoins)
	t.Log(utils.ToJson(rankedCoins))
}

func TestConvertSortResponseToCache(t *testing.T) {
	tests := []struct {
		name     string
		input    []dto.IntelligenceTokenRankResp
		expected []dto_cache.IntelligenceToken
	}{
		{
			name: "外部API数据转换测试",
			input: []dto.IntelligenceTokenRankResp{
				{
					ContractAddress:  "95KycpufBV37vuceYuwuNdyBV2a1HXu94ceEv6eUpump",
					Chain:            dto.ChainInfo{},
					ChainID:          0,
					Decimals:         6,
					Logo:             "",
					Name:             "American Bitcoin",
					Network:          "solana",
					PriceUSD:         "4.5295194983839864e-05",
					Symbol:           "ABTC",
					TotalSupply:      "",
					Volume24h:        "14256.278341192165",
					IsInternal:       false,
					Liquidity:        "13769.771650744993",
					CurrentMarketCap: "45184.689723956646",
				},
				{
					ContractAddress:  "3SmtvPSgYUS8HWFN7DiiGTA733r8QXim3MV1idJipump",
					Chain:            dto.ChainInfo{},
					ChainID:          0,
					Decimals:         6,
					Logo:             "",
					Name:             "American Bitcoin",
					Network:          "solana",
					PriceUSD:         "1.4474646916910041e-05",
					Symbol:           "ABTC",
					TotalSupply:      "",
					Volume24h:        "1911.1736697801687",
					IsInternal:       false,
					Liquidity:        "7626.083573756391",
					CurrentMarketCap: "14474.64691691004",
				},
			},
			expected: []dto_cache.IntelligenceToken{
				{
					ID:              "",
					EntityID:        "",
					Name:            "American Bitcoin",
					Symbol:          "ABTC",
					Standard:        nil,
					Decimals:        6,
					ContractAddress: "95KycpufBV37vuceYuwuNdyBV2a1HXu94ceEv6eUpump",
					Logo:            "",
					Stats: dto_cache.CoinMarketStats{
						WarningPriceUSD:     "0",
						WarningMarketCap:    "0",
						CurrentPriceUSD:     "4.5295194983839864e-05",
						CurrentMarketCap:    "45184.689723956646",
						HighestIncreaseRate: "0",
					},
					Chain: dto_cache.ChainInfo{
						ID:        "",
						NetworkID: "",
						Name:      "",
						Symbol:    "",
						Slug:      "solana",
						Logo:      "",
					},
					CreatedAt: dto_cache.CustomTime{},
					UpdatedAt: dto_cache.CustomTime{},
				},
				{
					ID:              "",
					EntityID:        "",
					Name:            "American Bitcoin",
					Symbol:          "ABTC",
					Standard:        nil,
					Decimals:        6,
					ContractAddress: "3SmtvPSgYUS8HWFN7DiiGTA733r8QXim3MV1idJipump",
					Logo:            "",
					Stats: dto_cache.CoinMarketStats{
						WarningPriceUSD:     "0",
						WarningMarketCap:    "0",
						CurrentPriceUSD:     "1.4474646916910041e-05",
						CurrentMarketCap:    "14474.64691691004",
						HighestIncreaseRate: "0",
					},
					Chain: dto_cache.ChainInfo{
						ID:        "",
						NetworkID: "",
						Name:      "",
						Symbol:    "",
						Slug:      "solana",
						Logo:      "",
					},
					CreatedAt: dto_cache.CustomTime{},
					UpdatedAt: dto_cache.CustomTime{},
				},
			},
		},
		{
			name: "内部数据转换测试",
			input: []dto.IntelligenceTokenRankResp{
				{
					ID:              "0198f631-3330-7e88-8608-a65b8fe48d37",
					EntityID:        "0197a526-3bbf-7596-9e72-dc716b9dc3df",
					Name:            "American bitcoin",
					Symbol:          "American bitcoin",
					Standard:        nil,
					Decimals:        18,
					ContractAddress: "0x4caa35c26d34297252f695bccf7908818507b949",
					Logo:            "",
					Chain: dto.ChainInfo{
						ID:        "019782be-e551-78b8-8582-62a47fa81f77",
						NetworkID: "",
						Name:      "Base",
						Symbol:    "",
						Slug:      "Base",
						Logo:      "assets/chain/base.png",
					},
					CreatedAt: "2005-08-29T14:17:56.273",
					UpdatedAt: "2005-09-02T18:14:38.681",
					Stats: dto.CoinMarketStats{
						WarningPriceUSD:     "0",
						WarningMarketCap:    "0",
						CurrentPriceUSD:     "0",
						CurrentMarketCap:    "0",
						HighestIncreaseRate: "0",
					},
				},
			},
			expected: []dto_cache.IntelligenceToken{
				{
					ID:              "0198f631-3330-7e88-8608-a65b8fe48d37",
					EntityID:        "0197a526-3bbf-7596-9e72-dc716b9dc3df",
					Name:            "American bitcoin",
					Symbol:          "American bitcoin",
					Standard:        nil,
					Decimals:        18,
					ContractAddress: "0x4caa35c26d34297252f695bccf7908818507b949",
					Logo:            "",
					Stats: dto_cache.CoinMarketStats{
						WarningPriceUSD:     "0",
						WarningMarketCap:    "0",
						CurrentPriceUSD:     "0",
						CurrentMarketCap:    "0",
						HighestIncreaseRate: "0",
					},
					Chain: dto_cache.ChainInfo{
						ID:        "019782be-e551-78b8-8582-62a47fa81f77",
						NetworkID: "",
						Name:      "Base",
						Symbol:    "",
						Slug:      "Base",
						Logo:      "assets/chain/base.png",
					},
					CreatedAt: dto_cache.CustomTime{Time: time.Date(2005, 8, 29, 14, 17, 56, 273000000, time.UTC)},
					UpdatedAt: dto_cache.CustomTime{Time: time.Date(2005, 9, 2, 18, 14, 38, 681000000, time.UTC)},
				},
			},
		},
		{
			name: "混合数据转换测试",
			input: []dto.IntelligenceTokenRankResp{
				// 外部API数据
				{
					ContractAddress:  "95KycpufBV37vuceYuwuNdyBV2a1HXu94ceEv6eUpump",
					Name:             "External Token",
					Network:          "ethereum",
					PriceUSD:         "1.0",
					CurrentMarketCap: "1000000",
				},
				// 内部数据
				{
					ID:              "internal-id",
					Name:            "Internal Token",
					ContractAddress: "0x123",
					Chain: dto.ChainInfo{
						Name: "Ethereum",
						Slug: "ethereum",
					},
					Stats: dto.CoinMarketStats{
						CurrentPriceUSD: "2.0",
					},
				},
			},
			expected: []dto_cache.IntelligenceToken{
				{
					ID:              "",
					EntityID:        "",
					Name:            "External Token",
					Symbol:          "",
					Standard:        nil,
					Decimals:        0,
					ContractAddress: "95KycpufBV37vuceYuwuNdyBV2a1HXu94ceEv6eUpump",
					Logo:            "",
					Stats: dto_cache.CoinMarketStats{
						WarningPriceUSD:     "0",
						WarningMarketCap:    "0",
						CurrentPriceUSD:     "1.0",
						CurrentMarketCap:    "1000000",
						HighestIncreaseRate: "0",
					},
					Chain: dto_cache.ChainInfo{
						ID:        "",
						NetworkID: "",
						Name:      "",
						Symbol:    "",
						Slug:      "ethereum",
						Logo:      "",
					},
					CreatedAt: dto_cache.CustomTime{},
					UpdatedAt: dto_cache.CustomTime{},
				},
				{
					ID:              "internal-id",
					EntityID:        "",
					Name:            "Internal Token",
					Symbol:          "",
					Standard:        nil,
					Decimals:        0,
					ContractAddress: "0x123",
					Logo:            "",
					Stats: dto_cache.CoinMarketStats{
						WarningPriceUSD:     "",
						WarningMarketCap:    "",
						CurrentPriceUSD:     "2.0",
						CurrentMarketCap:    "",
						HighestIncreaseRate: "",
					},
					Chain: dto_cache.ChainInfo{
						ID:        "",
						NetworkID: "",
						Name:      "Ethereum",
						Symbol:    "",
						Slug:      "ethereum",
						Logo:      "",
					},
					CreatedAt: dto_cache.CustomTime{},
					UpdatedAt: dto_cache.CustomTime{},
				},
			},
		},
		{
			name:     "空数据测试",
			input:    []dto.IntelligenceTokenRankResp{},
			expected: []dto_cache.IntelligenceToken{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertSortResponseToCache(tt.input)

			// 验证结果长度
			assert.Equal(t, len(tt.expected), len(result), "结果长度不匹配")

			// 验证每个字段
			for i, expected := range tt.expected {
				if i < len(result) {
					actual := result[i]

					// 验证基本信息
					assert.Equal(t, expected.ID, actual.ID, "ID不匹配")
					assert.Equal(t, expected.Name, actual.Name, "Name不匹配")
					assert.Equal(t, expected.Symbol, actual.Symbol, "Symbol不匹配")
					assert.Equal(t, expected.ContractAddress, actual.ContractAddress, "ContractAddress不匹配")
					assert.Equal(t, expected.Decimals, actual.Decimals, "Decimals不匹配")

					// 验证Stats
					assert.Equal(t, expected.Stats.CurrentPriceUSD, actual.Stats.CurrentPriceUSD, "CurrentPriceUSD不匹配")
					assert.Equal(t, expected.Stats.CurrentMarketCap, actual.Stats.CurrentMarketCap, "CurrentMarketCap不匹配")

					// 验证Chain
					assert.Equal(t, expected.Chain.Name, actual.Chain.Name, "Chain.Name不匹配")
					assert.Equal(t, expected.Chain.Slug, actual.Chain.Slug, "Chain.Slug不匹配")

					// 验证时间字段（如果有的话）
					if !expected.CreatedAt.Time.IsZero() {
						assert.Equal(t, expected.CreatedAt.Time.Format("2006-01-02T15:04:05"),
							actual.CreatedAt.Time.Format("2006-01-02T15:04:05"), "CreatedAt不匹配")
					}
					if !expected.UpdatedAt.Time.IsZero() {
						assert.Equal(t, expected.UpdatedAt.Time.Format("2006-01-02T15:04:05"),
							actual.UpdatedAt.Time.Format("2006-01-02T15:04:05"), "UpdatedAt不匹配")
					}
				}
			}
		})
	}
}

// 测试边界情况
func TestConvertSortResponseToCacheEdgeCases(t *testing.T) {
	t.Run("空字符串字段处理", func(t *testing.T) {
		input := []dto.IntelligenceTokenRankResp{
			{
				ContractAddress:  "",
				Name:             "",
				Network:          "",
				PriceUSD:         "",
				CurrentMarketCap: "",
			},
		}

		result := ConvertSortResponseToCache(input)
		assert.Len(t, result, 1)
		assert.Equal(t, "", result[0].Name)
		assert.Equal(t, "", result[0].ContractAddress)
	})

	t.Run("零值字段处理", func(t *testing.T) {
		input := []dto.IntelligenceTokenRankResp{
			{
				Decimals:   0,
				ChainID:    0,
				IsInternal: false,
			},
		}

		result := ConvertSortResponseToCache(input)
		assert.Len(t, result, 1)
		assert.Equal(t, 0, result[0].Decimals)
	})
}

// 测试ContractAddressAlt字段映射
func TestContractAddressAltMapping(t *testing.T) {
	// 模拟外部API返回的JSON数据
	jsonData := `[
		{
			"contractAddress": "95KycpufBV37vuceYuwuNdyBV2a1HXu94ceEv6eUpump",
			"name": "American Bitcoin",
			"network": "solana",
			"price_usd": "4.5295194983839864e-05",
			"current_market_cap": "45184.689723956646"
		}
	]`

	var tokens []dto.IntelligenceTokenRankResp
	err := json.Unmarshal([]byte(jsonData), &tokens)
	assert.NoError(t, err)
	assert.Len(t, tokens, 1)

	// 验证ContractAddressAlt字段是否正确设置
	token := tokens[0]
	t.Logf("Token after unmarshal: %+v", token)
	t.Logf("ContractAddress: '%s'", token.ContractAddress)
	t.Logf("ContractAddressAlt: '%s'", token.ContractAddressAlt)
	t.Logf("Network: '%s'", token.Network)

	// 验证字段映射
	assert.Equal(t, "95KycpufBV37vuceYuwuNdyBV2a1HXu94ceEv6eUpump", token.ContractAddressAlt)
	assert.Equal(t, "solana", token.Network)
	assert.Equal(t, "American Bitcoin", token.Name)

	// 现在测试ConvertSortResponseToCache函数
	result := ConvertSortResponseToCache(tokens)
	assert.Len(t, result, 1)

	cacheToken := result[0]
	t.Logf("Cache token: %+v", cacheToken)

	// 验证转换结果
	assert.Equal(t, "95KycpufBV37vuceYuwuNdyBV2a1HXu94ceEv6eUpump", cacheToken.ContractAddress)
	assert.Equal(t, "solana", cacheToken.Chain.Name)
	assert.Equal(t, "solana", cacheToken.Chain.Slug)
}

var originStr = `
[ {
  "contractAddress" : "95KycpufBV37vuceYuwuNdyBV2a1HXu94ceEv6eUpump",
  "chain" : "",
  "chain_id" : 0,
  "decimals" : 6,
  "logo" : "",
  "name" : "American Bitcoin",
  "network" : "solana",
  "price_usd" : "4.5295194983839864e-05",
  "symbol" : "ABTC",
  "total_supply" : "",
  "volume_24h" : "12368.180295846065",
  "is_internal" : false,
  "liquidity" : "13687.44734254751",
  "current_market_cap" : "45184.689723956646"
}, {
  "contractAddress" : "3SmtvPSgYUS8HWFN7DiiGTA733r8QXim3MV1idJipump",
  "chain" : "",
  "chain_id" : 0,
  "decimals" : 6,
  "logo" : "",
  "name" : "American Bitcoin",
  "network" : "solana",
  "price_usd" : "1.4474646916910041e-05",
  "symbol" : "ABTC",
  "total_supply" : "",
  "volume_24h" : "1742.393348398928",
  "is_internal" : false,
  "liquidity" : "7580.4900758837",
  "current_market_cap" : "14474.64691691004"
}, {
  "id" : "0198f631-3330-7e88-8608-a65b8fe48d37",
  "entity_id" : "0197a526-3bbf-7596-9e72-dc716b9dc3df",
  "name" : "American bitcoin",
  "symbol" : "American bitcoin",
  "standard" : null,
  "decimals" : 18,
  "contract_address" : "0x4caa35c26d34297252f695bccf7908818507b949",
  "logo" : "",
  "chain" : {
    "id" : "019782be-e551-78b8-8582-62a47fa81f77",
    "network_id" : "",
    "name" : "Base",
    "symbol" : "",
    "slug" : "Base",
    "logo" : "assets/chain/base.png"
  },
  "created_at" : "2025-08-29T14:17:56.273",
  "updated_at" : "2025-08-29T14:17:56.282",
  "stats" : {
    "warning_price_usd" : "0",
    "warning_market_cap" : "0",
    "current_price_usd" : "0",
    "current_market_cap" : "0",
    "highest_increase_rate" : "0"
  }
}, {
  "id" : "0198f631-3330-7e88-8608-a65b8fe48d37",
  "entity_id" : "0197a526-3bbf-7596-9e72-dc716b9dc3df",
  "name" : "American bitcoin",
  "symbol" : "American bitcoin",
  "standard" : null,
  "decimals" : 18,
  "contract_address" : "0x4caa35c26d34297252f695bccf7908818507b949",
  "logo" : "",
  "chain" : {
    "id" : "019782be-e551-78b8-8582-62a47fa81f77",
    "network_id" : "",
    "name" : "American bitcoin",
    "symbol" : "",
    "slug" : "",
    "logo" : ""
  },
  "created_at" : "2025-08-29T14:17:56.273",
  "updated_at" : "2025-09-02T19:40:07.846",
  "stats" : {
    "warning_price_usd" : "",
    "warning_market_cap" : "",
    "current_price_usd" : "",
    "current_market_cap" : "",
    "highest_increase_rate" : ""
  }
} ]
`

func TestUnmarshalJSON(t *testing.T) {
	lr.Init()
	var tokenList []dto.IntelligenceTokenRankResp
	if err := jsoniter.Unmarshal([]byte(originStr), &tokenList); err != nil {
		lr.E().Errorf("Failed to unmarshal data array: %v", err)
	}
	t.Log(utils.ToJson(tokenList))
}

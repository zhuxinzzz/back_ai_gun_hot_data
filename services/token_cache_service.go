package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/pkg/model/remote"
	"back_ai_gun_data/services/remote_service"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	IntelligenceCoinCacheKeyPrefix = "dogex:intelligence:latest_entities:intelligence_id:"
)

var chainName = map[string]struct{}{
	"Base":             {},
	"Polygon zkEVM":    {},
	"Fantom":           {},
	"Blast":            {},
	"Arbitrum":         {},
	"Scroll":           {},
	"Gnosis":           {},
	"Avalanche":        {},
	"Polygon":          {},
	"BSC":              {},
	"Optimism":         {},
	"Solana":           {},
	"zkSync":           {},
	"Linea":            {},
	"Moonbeam":         {},
	"Metis":            {},
	"Immutable zkEVM":  {},
	"Lisk":             {},
	"Soneium":          {},
	"Fuse":             {},
	"Lens":             {},
	"Mode":             {},
	"Gravity":          {},
	"Cronos":           {},
	"Abstract":         {},
	"Taiko":            {},
	"Boba":             {},
	"Unichain":         {},
	"opBNB":            {},
	"Swellchain":       {},
	"Aurora":           {},
	"Sei":              {},
	"Sonic":            {},
	"Moonriver":        {},
	"Corn":             {},
	"zkfair":           {},
	"Bitcoin":          {},
	"Conflux":          {},
	"Celo":             {},
	"Sui":              {},
	"binancecoin":      {},
	"World Chain":      {},
	"the-open-network": {},
	//"Ink":                     {},
	"BOB":                     {},
	"Superposition":           {},
	"kucoin-community-chain":  {},
	"arbitrum-nova":           {},
	"XDC":                     {},
	"Apechain":                {},
	"Kaia":                    {},
	"Avalanche X-Chain":       {},
	"Etherlink":               {},
	"BNB Beacon Chain (BEP2)": {},
	"Starknet":                {},
	"Rootstock":               {},
	"Berachain":               {},
	"Manta Pacific":           {},
	"Mantle":                  {},
	"Ethereum":                {},
	"HyperEVM":                {},
}

func UpdateTokenMarketData(ctx context.Context, intelligenceID string) error {
	cacheData, err := readTokenCache(ctx, intelligenceID)
	if err != nil {
		lr.E().Errorf("Failed to read intelligence coin cache: %v", err)
		return fmt.Errorf("failed to read intelligence coin cache: %w", err)
	}

	if len(cacheData) == 0 {
		lr.E().Errorf("No coins found in cache for intelligence %s", intelligenceID)
		return nil
	}

	// 批量获取GMGN数据
	updatedCount := 0

	// 收集所有需要查询的币种信息
	var queryParams []struct {
		index int
		coin  dto.IntelligenceCoinCache
	}

	for i, coin := range cacheData {
		if coin.Name != "" {
			queryParams = append(queryParams, struct {
				index int
				coin  dto.IntelligenceCoinCache
			}{i, coin})
		}
	}

	if len(queryParams) == 0 {
		lr.E().Errorf("No valid coins to query for intelligence %s", intelligenceID)
		return nil
	}

	// 按链分组，批量查询
	chainGroups := make(map[string][]struct {
		index int
		coin  dto.IntelligenceCoinCache
	})

	for _, param := range queryParams {
		chainSlug := param.coin.Chain.Slug
		if chainSlug == "" {
			chainSlug = "eth" // 默认链
		}
		chainGroups[chainSlug] = append(chainGroups[chainSlug], param)
	}

	// 为每个链批量查询
	for chainSlug, coins := range chainGroups {
		// 收集该链下所有币的名称
		var coinNames []string
		for _, param := range coins {
			coinNames = append(coinNames, param.coin.Name)
		}

		// 批量查询该链下的所有币
		namesStr := strings.Join(coinNames, ",")
		// 根据币种数量动态设置Limit，每个币平均3个结果
		limit := len(coinNames) * 3
		if limit < 10 {
			limit = 10 // 最少10个
		}

		// 过滤掉不支持的链
		if _, exists := chainName[chainSlug]; !exists {
			chainSlug = ""
		}
		tokens, err := remote_service.QueryTokensByNameWithLimit(ctx, namesStr, chainSlug, limit)
		if err != nil {
			lr.E().Errorf("Failed to batch query GMGN for chain %s: %v", chainSlug, err)
			continue
		}

		// 创建token映射，用于快速查找
		tokenMap := make(map[string]remote.GmGnToken)
		for _, token := range tokens {
			tokenMap[token.Name] = token
		}

		// 更新每个币的市场信息
		for _, param := range coins {
			coin := param.coin
			index := param.index

			// 优先使用合约地址匹配，如果没有合约地址则使用名称匹配
			var matchedToken *remote.GmGnToken

			if coin.ContractAddress != "" {
				// 使用合约地址精确匹配
				for _, token := range tokens {
					if strings.EqualFold(token.Address, coin.ContractAddress) {
						matchedToken = &token
						break
					}
				}
			}

			// 如果合约地址匹配失败，使用名称匹配
			if matchedToken == nil {
				if token, exists := tokenMap[coin.Name]; exists {
					matchedToken = &token
				}
			}

			// 更新市场信息
			if matchedToken != nil {
				cacheData[index].Stats.CurrentPriceUSD = matchedToken.PriceUSD
				cacheData[index].Stats.CurrentMarketCap = matchedToken.MarketCap
				cacheData[index].UpdatedAt.Time = time.Now()

				// 计算预警涨幅：当前市值 ÷ 预警市值
				if matchedToken.MarketCap != "0" && coin.Stats.WarningMarketCap != "0" {
					currentMarketCap, err1 := strconv.ParseFloat(matchedToken.MarketCap, 64)
					warningMarketCap, err2 := strconv.ParseFloat(coin.Stats.WarningMarketCap, 64)

					if err1 == nil && err2 == nil && warningMarketCap > 0 {
						currentIncreaseRate := currentMarketCap / warningMarketCap

						// 获取历史最高涨幅
						highestIncreaseRate, err3 := strconv.ParseFloat(coin.Stats.HighestIncreaseRate, 64)
						if err3 != nil {
							highestIncreaseRate = 0
						}

						// 更新最高涨幅（取较大值）
						if currentIncreaseRate > highestIncreaseRate {
							cacheData[index].Stats.HighestIncreaseRate = fmt.Sprintf("%.6f", currentIncreaseRate)
						}
					}
				}

				// 如果合约地址为空，更新合约地址
				if cacheData[index].ContractAddress == "" && matchedToken.Address != "" {
					cacheData[index].ContractAddress = matchedToken.Address
				}

				updatedCount++
			} else {
				lr.E().Errorf("No GMGN data found for coin: %s (chain: %s)", coin.Name, chainSlug)
			}
		}
	}

	// 将更新后的数据写回缓存
	if err := writeTokenCache(ctx, intelligenceID, cacheData); err != nil {
		lr.E().Errorf("Failed to write intelligence coin cache: %v", err)
		return fmt.Errorf("failed to write intelligence coin cache: %w", err)
	}

	//lr.I().Infof("Updated market data for intelligence %s: %d/%d coins", intelligenceID, updatedCount, len(cacheData))
	return nil
}

func ReadTokenCache(ctx context.Context, intelligenceID string) ([]dto.IntelligenceCoinCache, error) {
	return readTokenCache(ctx, intelligenceID)
}

func readTokenCache(ctx context.Context, intelligenceID string) ([]dto.IntelligenceCoinCache, error) {
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID

	cacheData, err := cache.Get(ctx, cacheKey)
	if err != nil {
		// 如果缓存不存在，返回空数据而不是错误
		if err.Error() == "redis: nil" {
			lr.E().Errorf("Cache not found for intelligence %s, returning empty data", intelligenceID)
			return []dto.IntelligenceCoinCache{}, nil
		}
		return nil, fmt.Errorf("failed to get cache data: %w", err)
	}

	// 直接解析为币数组
	var coins []dto.IntelligenceCoinCache
	if err := json.Unmarshal([]byte(cacheData), &coins); err != nil {
		lr.E().Errorf("Failed to unmarshal cache data: %v", err)
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return coins, nil
}

func writeTokenCache(ctx context.Context, intelligenceID string, coins []dto.IntelligenceCoinCache) error {
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID

	data, err := json.Marshal(coins)
	if err != nil {
		lr.E().Errorf("Failed to marshal intelligence coin cache: %v", err)
		return fmt.Errorf("failed to marshal intelligence coin cache: %w", err)
	}

	// 写入Redis缓存，设置过期时间为4天
	if err := cache.Set(ctx, cacheKey, string(data), 4*24*time.Hour); err != nil {
		lr.E().Errorf("Failed to write cache data: %v", err)
		return fmt.Errorf("failed to write cache data: %w", err)
	}

	return nil
}

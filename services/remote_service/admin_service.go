package remote_service

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/pkg/model/remote"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// 缓存键前缀
	IntelligenceCoinCacheKeyPrefix = "dogex:intelligence:latest_entity_info:intelligence_id:"
)

// UpdateAdminMarketData 更新admin服务缓存中的市场信息
// 从缓存读取数据，使用GMGN查询更新市场信息，然后写回缓存
func UpdateAdminMarketData(intelligenceID string) error {
	// 从缓存读取现有的币种数据
	cacheData, err := readIntelligenceCoinCacheFromRedis(intelligenceID)
	if err != nil {
		lr.E().Errorf("Failed to read intelligence coin cache: %v", err)
		return fmt.Errorf("failed to read intelligence coin cache: %w", err)
	}

	if cacheData == nil || len(cacheData.Coins) == 0 {
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

	for i, coin := range cacheData.Coins {
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
		tokens, err := QueryTokensByName(namesStr, chainSlug)
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
				cacheData.Coins[index].Stats.CurrentPriceUSD = matchedToken.PriceUSD
				cacheData.Coins[index].Stats.CurrentMarketCap = matchedToken.MarketCap
				cacheData.Coins[index].UpdatedAt = time.Now()

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
							cacheData.Coins[index].Stats.HighestIncreaseRate = fmt.Sprintf("%.6f", currentIncreaseRate)
						}
					}
				}

				// 如果合约地址为空，更新合约地址
				if cacheData.Coins[index].ContractAddress == "" && matchedToken.Address != "" {
					cacheData.Coins[index].ContractAddress = matchedToken.Address
				}

				updatedCount++
			} else {
				lr.E().Errorf("No GMGN data found for coin: %s (chain: %s)", coin.Name, chainSlug)
			}
		}
	}

	// 更新缓存时间戳
	cacheData.UpdatedAt = time.Now()

	// 将更新后的数据写回缓存
	if err := writeIntelligenceCoinCacheToRedis(intelligenceID, cacheData.Coins); err != nil {
		lr.E().Errorf("Failed to write intelligence coin cache: %v", err)
		return fmt.Errorf("failed to write intelligence coin cache: %w", err)
	}

	lr.I().Infof("Updated market data for intelligence %s: %d/%d coins", intelligenceID, updatedCount, len(cacheData.Coins))
	return nil
}

// ReadIntelligenceCoinCacheFromRedis 从Redis读取情报币缓存（公共接口）
func ReadIntelligenceCoinCacheFromRedis(intelligenceID string) (*dto.IntelligenceCoinCacheData, error) {
	return readIntelligenceCoinCacheFromRedis(intelligenceID)
}

// CallAdminRankingService 调用admin服务的排序接口
func CallAdminRankingService(coins []dto.IntelligenceCoinCache) (*dto.AdminRankingResponse, error) {
	// 模拟admin服务排序接口
	// 接口：POST /api/admin/ranking
	// Body: 代币集合
	// 返回：排序后的代币集合（按市值降序排列）

	if len(coins) == 0 {
		return &dto.AdminRankingResponse{
			Code:    0,
			Message: "success",
			Data:    []dto.IntelligenceCoinCache{},
		}, nil
	}

	// 复制一份数据进行排序，避免修改原数据
	rankedCoins := make([]dto.IntelligenceCoinCache, len(coins))
	copy(rankedCoins, coins)

	// 按市值降序排序（模拟admin服务的排序逻辑）
	// 优先使用current_market_cap，如果为0则使用warning_market_cap
	sort.Slice(rankedCoins, func(i, j int) bool {
		// 获取i的市值
		iMarketCap := rankedCoins[i].Stats.CurrentMarketCap
		if iMarketCap == "0" || iMarketCap == "" {
			iMarketCap = rankedCoins[i].Stats.WarningMarketCap
		}

		// 获取j的市值
		jMarketCap := rankedCoins[j].Stats.CurrentMarketCap
		if jMarketCap == "0" || jMarketCap == "" {
			jMarketCap = rankedCoins[j].Stats.WarningMarketCap
		}

		// 转换为float进行比较
		iVal, err1 := strconv.ParseFloat(iMarketCap, 64)
		jVal, err2 := strconv.ParseFloat(jMarketCap, 64)

		// 如果解析失败，按字符串比较
		if err1 != nil || err2 != nil {
			return iMarketCap > jMarketCap
		}

		// 按数值降序排列
		return iVal > jVal
	})

	// 模拟网络延迟
	time.Sleep(10 * time.Millisecond)

	// 返回排序后的数据
	response := &dto.AdminRankingResponse{
		Code:    0,
		Message: "success",
		Data:    rankedCoins,
	}

	return response, nil
}

// readIntelligenceCoinCacheFromRedis 从Redis读取情报币缓存
func readIntelligenceCoinCacheFromRedis(intelligenceID string) (*dto.IntelligenceCoinCacheData, error) {
	ctx := context.Background()
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID

	cacheData, err := cache.Get(ctx, cacheKey)
	if err != nil {
		// 如果缓存不存在，返回空数据而不是错误
		if err.Error() == "redis: nil" {
			lr.E().Errorf("Cache not found for intelligence %s, returning empty data", intelligenceID)
			return &dto.IntelligenceCoinCacheData{
				IntelligenceID: intelligenceID,
				Coins:          []dto.IntelligenceCoinCache{},
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get cache data: %w", err)
	}

	var data dto.IntelligenceCoinCacheData
	if err := json.Unmarshal([]byte(cacheData), &data); err != nil {
		lr.E().Errorf("Failed to unmarshal cache data: %v", err)
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return &data, nil
}

// writeIntelligenceCoinCacheToRedis 将情报币缓存写入Redis
func writeIntelligenceCoinCacheToRedis(intelligenceID string, coins []dto.IntelligenceCoinCache) error {
	ctx := context.Background()
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID

	// 构建缓存数据结构
	cacheData := dto.IntelligenceCoinCacheData{
		IntelligenceID: intelligenceID,
		Coins:          coins,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 序列化数据
	data, err := json.Marshal(cacheData)
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

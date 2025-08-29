package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/services/remote_service"
	"back_ai_gun_data/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"back_ai_gun_data/pkg/lr"
)

const (
	// 币热数据缓存键前缀
	CoinHotDataCacheKeyPrefix = "dogex:coin:hot_data:"

	// 币热数据缓存过期时间：7天
	CoinHotDataCacheExpiration = 7 * 24 * time.Hour
)

// ProcessCoinHotData 处理币热数据流程
// 1. 更新admin服务缓存中的市场信息
// 2. 调用admin服务排序接口
// 3. 打热点标签（is_show字段）
// 4. 进入热数据缓存
func ProcessCoinHotData(coins []dto.IntelligenceCoinCache) error {
	lr.I().Infof("Processing coin hot data for %d coins", len(coins))

	// 步骤1: 更新admin服务缓存中的市场信息
	if err := updateAdminMarketData(coins); err != nil {
		lr.E().Errorf("Failed to update admin market data: %v", err)
		// 继续执行，不中断流程
	}

	// 步骤2: 调用admin服务排序接口
	rankingResponse, err := callAdminRankingService(coins)
	if err != nil {
		lr.E().Errorf("Failed to call admin ranking service: %v", err)
		return err
	}

	// 步骤3: 打热点标签并进入热数据缓存
	if err := processHotDataLabels(rankingResponse.Data); err != nil {
		lr.E().Errorf("Failed to process hot data labels: %v", err)
		return err
	}

	lr.I().Infof("Successfully processed coin hot data for %d coins", len(coins))
	return nil
}

// updateAdminMarketData 更新admin服务缓存中的市场信息
func updateAdminMarketData(coins []dto.IntelligenceCoinCache) error {
	var marketData []dto.AdminMarketData

	for _, coin := range coins {
		// 跳过没有名称或市场数据的币
		if coin.Name == "" || (coin.Stats.CurrentPriceUSD == "" && coin.Stats.CurrentMarketCap == "") {
			continue
		}

		adminData := dto.AdminMarketData{
			CoinID:           coin.ID,
			Name:             coin.Name,
			Symbol:           coin.Symbol,
			ContractAddress:  coin.ContractAddress, // 可能为空，但不影响主要功能
			Chain:            coin.Chain.Slug,
			CurrentPriceUSD:  coin.Stats.CurrentPriceUSD,
			CurrentMarketCap: coin.Stats.CurrentMarketCap,
			Ranking:          0, // 初始排名为0
			IsShow:           false,
			UpdatedAt:        time.Now().Format(time.RFC3339),
		}
		marketData = append(marketData, adminData)
	}

	if len(marketData) == 0 {
		lr.I().Info("No market data to update")
		return nil
	}

	// 调用admin服务更新市场信息
	if err := remote_service.UpdateAdminMarketData(marketData); err != nil {
		return fmt.Errorf("failed to update admin market data: %w", err)
	}

	lr.I().Infof("Updated admin market data for %d coins", len(marketData))
	return nil
}

// callAdminRankingService 调用admin服务排序接口
func callAdminRankingService(coins []dto.IntelligenceCoinCache) (*dto.AdminRankingResponse, error) {
	var adminCoins []dto.AdminMarketData

	for _, coin := range coins {
		// 跳过没有名称或市场数据的币
		if coin.Name == "" || (coin.Stats.CurrentPriceUSD == "" && coin.Stats.CurrentMarketCap == "") {
			continue
		}

		adminData := dto.AdminMarketData{
			CoinID:           coin.ID,
			Name:             coin.Name,
			Symbol:           coin.Symbol,
			ContractAddress:  coin.ContractAddress, // 可能为空，但不影响主要功能
			Chain:            coin.Chain.Slug,
			CurrentPriceUSD:  coin.Stats.CurrentPriceUSD,
			CurrentMarketCap: coin.Stats.CurrentMarketCap,
			Ranking:          0,
			IsShow:           false,
			UpdatedAt:        time.Now().Format(time.RFC3339),
		}
		adminCoins = append(adminCoins, adminData)
	}

	if len(adminCoins) == 0 {
		lr.I().Info("No coins to rank")
		return &dto.AdminRankingResponse{
			Code:    0,
			Message: "success",
			Data:    []dto.AdminMarketData{},
		}, nil
	}

	// 调用admin服务排序接口
	response, err := remote_service.CallAdminRankingService(adminCoins)
	if err != nil {
		return nil, fmt.Errorf("failed to call admin ranking service: %w", err)
	}

	lr.I().Infof("Called admin ranking service for %d coins", len(adminCoins))
	return response, nil
}

// processHotDataLabels 处理热点标签并进入热数据缓存
func processHotDataLabels(rankedCoins []dto.AdminMarketData) error {
	var hotDataCoins []dto.CoinHotData

	for _, rankedCoin := range rankedCoins {
		// 检查是否进入前三名
		isTopThree := rankedCoin.Ranking > 0 && rankedCoin.Ranking <= 3

		// 创建热数据
		hotData := dto.CoinHotData{
			ID:              rankedCoin.CoinID,
			EntityID:        utils.GenerateUUIDV7(),
			Name:            rankedCoin.Name,
			Symbol:          rankedCoin.Symbol,
			Standard:        stringPtr("ERC20"),
			Decimals:        18,
			ContractAddress: rankedCoin.ContractAddress,
			Logo:            "",
			Stats: dto.CoinMarketStats{
				CurrentPriceUSD:  rankedCoin.CurrentPriceUSD,
				CurrentMarketCap: rankedCoin.CurrentMarketCap,
			},
			Chain: dto.ChainInfo{
				Slug: rankedCoin.Chain,
			},
			IsShow:         rankedCoin.IsShow || isTopThree,
			Ranking:        rankedCoin.Ranking,
			HighestRanking: rankedCoin.Ranking,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// 如果是首次进入前三，记录时间
		if isTopThree {
			now := time.Now()
			hotData.FirstRankedAt = &now
			hotData.LastRankedAt = &now
		}

		// 只有被打上is_show标签的币才进入热数据缓存
		if hotData.IsShow {
			hotDataCoins = append(hotDataCoins, hotData)
			lr.I().Infof("Added coin %s to hot data cache (ranking: %d)", hotData.Name, hotData.Ranking)
		}
	}

	// 更新热数据缓存
	if len(hotDataCoins) > 0 {
		if err := updateCoinHotDataCache(hotDataCoins); err != nil {
			return fmt.Errorf("failed to update coin hot data cache: %w", err)
		}
		lr.I().Infof("Updated coin hot data cache with %d coins", len(hotDataCoins))
	}

	return nil
}

// updateCoinHotDataCache 更新币热数据缓存
func updateCoinHotDataCache(hotDataCoins []dto.CoinHotData) error {
	// 获取现有的热数据缓存
	existingCache, err := getCoinHotDataCache()
	if err != nil {
		lr.E().Errorf("Failed to get existing hot data cache: %v", err)
		// 创建新的缓存
		existingCache = &dto.CoinHotDataCache{
			Coins:     []dto.CoinHotData{},
			UpdatedAt: time.Now(),
		}
	}

	// 创建币ID映射，用于快速查找
	existingCoinsMap := make(map[string]dto.CoinHotData)
	for _, coin := range existingCache.Coins {
		existingCoinsMap[coin.ID] = coin
	}

	// 更新或添加新的热数据
	for _, newCoin := range hotDataCoins {
		existingCoinsMap[newCoin.ID] = newCoin
	}

	// 转换回切片
	var updatedCoins []dto.CoinHotData
	for _, coin := range existingCoinsMap {
		updatedCoins = append(updatedCoins, coin)
	}

	// 更新缓存
	updatedCache := &dto.CoinHotDataCache{
		Coins:     updatedCoins,
		UpdatedAt: time.Now(),
	}

	return setCoinHotDataCache(updatedCache)
}

// getCoinHotDataCache 获取币热数据缓存
func getCoinHotDataCache() (*dto.CoinHotDataCache, error) {
	ctx := context.Background()
	cacheKey := CoinHotDataCacheKeyPrefix + "all"

	cacheData, err := cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var data dto.CoinHotDataCache
	if err := json.Unmarshal([]byte(cacheData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hot data cache: %w", err)
	}

	return &data, nil
}

// setCoinHotDataCache 设置币热数据缓存
func setCoinHotDataCache(data *dto.CoinHotDataCache) error {
	ctx := context.Background()
	cacheKey := CoinHotDataCacheKeyPrefix + "all"

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal hot data cache: %w", err)
	}

	return cache.Set(ctx, cacheKey, string(jsonData), CoinHotDataCacheExpiration)
}

// GetCoinHotData 获取币热数据
func GetCoinHotData() ([]dto.CoinHotData, error) {
	cache, err := getCoinHotDataCache()
	if err != nil {
		return nil, err
	}
	return cache.Coins, nil
}

// DeleteCoinHotData 删除币热数据
func DeleteCoinHotData(coinID string) error {
	ctx := context.Background()
	cacheKey := CoinHotDataCacheKeyPrefix + coinID
	return cache.Del(ctx, cacheKey)
}

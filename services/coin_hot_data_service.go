package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/services/remote_service"
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
func ProcessCoinHotData(intelligenceID string, coins []dto.IntelligenceCoinCache) error {
	// 步骤1: 更新admin服务缓存中的市场信息
	if err := updateAdminMarketData(intelligenceID, coins); err != nil {
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

	return nil
}

// updateAdminMarketData 更新admin服务缓存中的市场信息
func updateAdminMarketData(intelligenceID string, coins []dto.IntelligenceCoinCache) error {
	// 调用admin服务更新市场信息
	if err := remote_service.UpdateAdminMarketData(nil, intelligenceID); err != nil {
		lr.E().Errorf("Failed to update admin market data: %v", err)
		return fmt.Errorf("failed to update admin market data: %w", err)
	}

	return nil
}

// callAdminRankingService 调用admin服务排序接口
func callAdminRankingService(coins []dto.IntelligenceCoinCache) (*dto.AdminRankingResponse, error) {
	// 过滤有效的币种
	var validCoins []dto.IntelligenceCoinCache
	for _, coin := range coins {
		// 跳过没有名称的币
		if coin.Name == "" {
			continue
		}
		validCoins = append(validCoins, coin)
	}

	if len(validCoins) == 0 {
		return &dto.AdminRankingResponse{
			Code:    0,
			Message: "success",
			Data:    []dto.IntelligenceCoinCache{},
		}, nil
	}

	// 调用admin服务排序接口
	response, err := remote_service.CallAdminRankingService(validCoins)
	if err != nil {
		lr.E().Errorf("Failed to call admin ranking service: %v", err)
		return nil, fmt.Errorf("failed to call admin ranking service: %w", err)
	}

	return response, nil
}

// processHotDataLabels 处理热点标签并进入热数据缓存
func processHotDataLabels(rankedCoins []dto.IntelligenceCoinCache) error {
	var hotDataCoins []dto.CoinHotData

	for i, rankedCoin := range rankedCoins {
		// 根据排序结果确定是否进入前三名
		// 假设rankedCoins已经按排名排序，前三个是前三名
		isTopThree := i < 3
		ranking := i + 1 // 排名从1开始

		// 创建热数据
		hotData := dto.CoinHotData{
			ID:              rankedCoin.ID,
			EntityID:        rankedCoin.EntityID,
			Name:            rankedCoin.Name,
			Symbol:          rankedCoin.Symbol,
			Standard:        rankedCoin.Standard,
			Decimals:        rankedCoin.Decimals,
			ContractAddress: rankedCoin.ContractAddress,
			Logo:            rankedCoin.Logo,
			Stats:           rankedCoin.Stats,
			Chain:           rankedCoin.Chain,
			IsShow:          isTopThree, // 前三名显示
			Ranking:         ranking,    // 当前排名
			HighestRanking:  ranking,    // 暂时设为当前排名，后续需要比较历史最高
			CreatedAt:       rankedCoin.CreatedAt,
			UpdatedAt:       time.Now(),
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
		}
	}

	// 更新热数据缓存
	if len(hotDataCoins) > 0 {
		if err := updateCoinHotDataCache(hotDataCoins); err != nil {
			lr.E().Errorf("Failed to update coin hot data cache: %v", err)
			return fmt.Errorf("failed to update coin hot data cache: %w", err)
		}
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
		lr.E().Errorf("Failed to unmarshal hot data cache: %v", err)
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
		lr.E().Errorf("Failed to marshal hot data cache: %v", err)
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

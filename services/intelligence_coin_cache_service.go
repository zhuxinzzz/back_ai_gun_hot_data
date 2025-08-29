package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/model"
	"back_ai_gun_data/pkg/model/dto"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/utils"
)

const (
	// 缓存键前缀
	IntelligenceCoinCacheKeyPrefix = "dogex:intelligence:latest_entity_info:intelligence_id:"

	// 缓存过期时间：4天
	CacheExpiration = 4 * 24 * time.Hour

	// 持久化触发时间：3天
	PersistenceTriggerTime = 3 * 24 * time.Hour
)

// ProcessIntelligenceCoinCache 处理情报-币缓存（主入口函数）
func ProcessIntelligenceCoinCache(data *model.MessageData) error {
	intelligenceID := data.ID

	// 从消息中提取币信息
	coins, err := extractCoinsFromMessage(data)
	if err != nil {
		return fmt.Errorf("failed to extract coins from message: %w", err)
	}

	if len(coins) == 0 {
		lr.I().Infof("No coins found in message %s", data.ID)
		return nil
	}

	// 更新缓存
	if err := updateIntelligenceCoinCache(intelligenceID, coins); err != nil {
		return fmt.Errorf("failed to update intelligence coin cache: %w", err)
	}

	lr.I().Infof("Successfully updated intelligence coin cache for intelligence %s with %d coins", intelligenceID, len(coins))
	return nil
}

// extractCoinsFromMessage 从消息中提取币信息
func extractCoinsFromMessage(data *model.MessageData) ([]dto.IntelligenceCoinCache, error) {
	var coins []dto.IntelligenceCoinCache

	// 从tokens中提取币信息
	for _, tokenName := range data.Data.EntitiesExtract.Entities.Tokens {
		coin := dto.IntelligenceCoinCache{
			ID:              utils.GenerateUUIDV7(), // 生成project chain data id
			EntityID:        utils.GenerateUUIDV7(), // 生成实体ID
			Name:            tokenName,
			Symbol:          generateSymbol(tokenName),
			Standard:        stringPtr("ERC20"), // 默认标准
			Decimals:        18,                 // 默认精度
			ContractAddress: "",                 // 暂时为空，后续从搜索中获取
			Logo:            "",                 // 暂时为空，后续从搜索中获取
			Stats: dto.CoinMarketStats{
				WarningPriceUSD:     "0",
				WarningMarketCap:    "0",
				CurrentPriceUSD:     "0",
				CurrentMarketCap:    "0",
				HighestIncreaseRate: "0",
			},
			Chain: dto.ChainInfo{
				ID:   "default-chain-id",
				Name: "Ethereum",
				Slug: "eth",
				Logo: "assets/chain/eth.png",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		coins = append(coins, coin)
	}

	// 从projects中提取币信息（如果有的话）
	for _, projectName := range data.Data.EntitiesExtract.Entities.Projects {
		coin := dto.IntelligenceCoinCache{
			ID:              utils.GenerateUUIDV7(),
			EntityID:        utils.GenerateUUIDV7(),
			Name:            projectName,
			Symbol:          generateSymbol(projectName),
			Standard:        stringPtr("ERC20"),
			Decimals:        18,
			ContractAddress: "",
			Logo:            "",
			Stats: dto.CoinMarketStats{
				WarningPriceUSD:     "0",
				WarningMarketCap:    "0",
				CurrentPriceUSD:     "0",
				CurrentMarketCap:    "0",
				HighestIncreaseRate: "0",
			},
			Chain: dto.ChainInfo{
				ID:   "default-chain-id",
				Name: "Ethereum",
				Slug: "eth",
				Logo: "assets/chain/eth.png",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		coins = append(coins, coin)
	}

	return coins, nil
}

// generateSymbol 生成币符号
func generateSymbol(name string) string {
	if len(name) >= 3 {
		return name[:3]
	}
	return name
}

// updateIntelligenceCoinCache 更新情报-币缓存
func updateIntelligenceCoinCache(intelligenceID string, newCoins []dto.IntelligenceCoinCache) error {
	// 获取现有缓存数据
	existingData, err := getIntelligenceCoinCache(intelligenceID)
	if err != nil {
		lr.E().Errorf("Failed to get existing cache for intelligence %s: %v", intelligenceID, err)
		// 如果获取失败，创建新的缓存数据
		existingData = &dto.IntelligenceCoinCacheData{
			IntelligenceID: intelligenceID,
			Coins:          []dto.IntelligenceCoinCache{},
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
	}

	// 合并币信息（避免重复）
	existingCoins := make(map[string]dto.IntelligenceCoinCache)
	for _, coin := range existingData.Coins {
		existingCoins[coin.Name] = coin
	}

	// 添加或更新新币信息
	for _, newCoin := range newCoins {
		if existingCoin, exists := existingCoins[newCoin.Name]; exists {
			// 更新现有币信息，保留原有的市场数据
			newCoin.Stats = existingCoin.Stats
			newCoin.UpdatedAt = time.Now()
		}
		existingCoins[newCoin.Name] = newCoin
	}

	// 转换回切片
	var updatedCoins []dto.IntelligenceCoinCache
	for _, coin := range existingCoins {
		updatedCoins = append(updatedCoins, coin)
	}

	// 更新缓存数据
	updatedData := &dto.IntelligenceCoinCacheData{
		IntelligenceID: intelligenceID,
		Coins:          updatedCoins,
		CreatedAt:      existingData.CreatedAt,
		UpdatedAt:      time.Now(),
	}

	// 检查是否需要持久化
	if shouldPersist(updatedData) {
		if err := persistIntelligenceCoinData(updatedData); err != nil {
			lr.E().Errorf("Failed to persist intelligence coin data for %s: %v", intelligenceID, err)
			// 持久化失败不影响缓存更新
		}
	}

	// 更新缓存
	return setIntelligenceCoinCache(intelligenceID, updatedData)
}

// getIntelligenceCoinCache 获取情报-币缓存
func getIntelligenceCoinCache(intelligenceID string) (*dto.IntelligenceCoinCacheData, error) {
	ctx := context.Background()
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID

	cacheData, err := cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var data dto.IntelligenceCoinCacheData
	if err := json.Unmarshal([]byte(cacheData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return &data, nil
}

// setIntelligenceCoinCache 设置情报-币缓存
func setIntelligenceCoinCache(intelligenceID string, data *dto.IntelligenceCoinCacheData) error {
	ctx := context.Background()
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	return cache.Set(ctx, cacheKey, string(jsonData), CacheExpiration)
}

// shouldPersist 判断是否需要持久化
func shouldPersist(data *dto.IntelligenceCoinCacheData) bool {
	// 如果缓存时间超过3天，触发持久化
	return time.Since(data.CreatedAt) >= PersistenceTriggerTime
}

// persistIntelligenceCoinData 持久化情报-币数据
func persistIntelligenceCoinData(data *dto.IntelligenceCoinCacheData) error {
	// TODO: 实现持久化逻辑
	// 这里可以存储到数据库、文件系统或其他持久化存储
	lr.I().Infof("Persisting intelligence coin data for intelligence %s with %d coins", data.IntelligenceID, len(data.Coins))

	// 示例：记录到日志
	for _, coin := range data.Coins {
		lr.I().Infof("Persisting coin: %s (ID: %s, EntityID: %s)", coin.Name, coin.ID, coin.EntityID)
	}

	return nil
}

// GetIntelligenceCoins 获取情报关联的币信息
func GetIntelligenceCoins(intelligenceID string) ([]dto.IntelligenceCoinCache, error) {
	data, err := getIntelligenceCoinCache(intelligenceID)
	if err != nil {
		return nil, err
	}
	return data.Coins, nil
}

// DeleteIntelligenceCoinCache 删除情报-币缓存
func DeleteIntelligenceCoinCache(intelligenceID string) error {
	ctx := context.Background()
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID
	return cache.Del(ctx, cacheKey)
}

// UpdateCoinMarketData 更新币的市场数据
func UpdateCoinMarketData(intelligenceID, coinName string, marketStats dto.CoinMarketStats) error {
	data, err := getIntelligenceCoinCache(intelligenceID)
	if err != nil {
		return err
	}

	// 查找并更新指定币的市场数据
	for i, coin := range data.Coins {
		if coin.Name == coinName {
			data.Coins[i].Stats = marketStats
			data.Coins[i].UpdatedAt = time.Now()
			data.UpdatedAt = time.Now()
			break
		}
	}

	return setIntelligenceCoinCache(intelligenceID, data)
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

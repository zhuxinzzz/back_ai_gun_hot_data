package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/model/dto_cache"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"back_ai_gun_data/pkg/lr"
)

const (
	// 币热数据缓存键前缀
	CoinHotDataCacheKeyPrefix = "dogex:token:hot_data:"

	// 币热数据缓存过期时间：7天
	CoinHotDataCacheExpiration = 7 * 24 * time.Hour
)

// ProcessCoinHotData 处理币热数据流程
// 输入的 coins 应为已按排名排序的列表
// 逻辑：仅当当前前三中出现“新成员”时，按顺序追加到缓存；已存在的历史成员保留
func ProcessCoinHotData(intelligenceID string, coins []dto_cache.IntelligenceTokenCache) error {
	if err := appendNewTopThree(coins); err != nil {
		lr.E().Errorf("Failed to process hot data: %v", err)
		return err
	}
	return nil
}

// appendNewTopThree 仅将当前排序结果的前三名里“新出现”的成员按顺序追加到缓存
func appendNewTopThree(rankedCoins []dto_cache.IntelligenceTokenCache) error {
	// 只关心当前排名的前三名
	topN := 3
	if len(rankedCoins) < topN {
		topN = len(rankedCoins)
	}
	if topN == 0 {
		return nil
	}

	// 读取现有热数据缓存（为一个全局列表）
	existing, err := getCoinHotDataCache()
	if err != nil || existing == nil {
		existing = []dto_cache.IntelligenceTokenCache{}
	}

	// 构建已存在ID集合，避免重复
	existingIDs := make(map[string]struct{}, len(existing))
	for _, c := range existing {
		existingIDs[c.ID] = struct{}{}
	}

	// 按顺序检查前三，如果是新成员则按顺序追加
	for i := 0; i < topN; i++ {
		rc := rankedCoins[i]
		if _, ok := existingIDs[rc.ID]; ok {
			continue
		}
		existing = append(existing, rc)
	}

	// 若没有新增则不写回
	if len(existingIDs) == len(existing) {
		return nil
	}

	return setCoinHotDataCache(existing)
}

// getCoinHotDataCache 获取币热数据缓存（直接返回代币集合）
func getCoinHotDataCache() ([]dto_cache.IntelligenceTokenCache, error) {
	ctx := context.Background()
	cacheKey := CoinHotDataCacheKeyPrefix + "all"

	cacheData, err := cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var data []dto_cache.IntelligenceTokenCache
	if err := json.Unmarshal([]byte(cacheData), &data); err != nil {
		lr.E().Errorf("Failed to unmarshal hot data cache: %v", err)
		return nil, fmt.Errorf("failed to unmarshal hot data cache: %w", err)
	}

	return data, nil
}

// setCoinHotDataCache 设置币热数据缓存（直接存储代币集合）
func setCoinHotDataCache(coins []dto_cache.IntelligenceTokenCache) error {
	ctx := context.Background()
	cacheKey := CoinHotDataCacheKeyPrefix + "all"

	jsonData, err := json.Marshal(coins)
	if err != nil {
		lr.E().Errorf("Failed to marshal hot data cache: %v", err)
		return fmt.Errorf("failed to marshal hot data cache: %w", err)
	}

	return cache.Set(ctx, cacheKey, string(jsonData), CoinHotDataCacheExpiration)
}

// GetCoinHotData 获取币热数据（直接返回代币集合）
func GetCoinHotData() ([]dto_cache.IntelligenceTokenCache, error) {
	data, err := getCoinHotDataCache()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DeleteCoinHotData 从热数据中移除指定coin ID（读-改-写）
func DeleteCoinHotData(coinID string) error {
	ctx := context.Background()
	cacheKey := CoinHotDataCacheKeyPrefix + "all"

	existing, err := getCoinHotDataCache()
	if err != nil {
		return err
	}

	var filtered []dto_cache.IntelligenceTokenCache
	for _, c := range existing {
		if c.ID != coinID {
			filtered = append(filtered, c)
		}
	}

	jsonData, err := json.Marshal(filtered)
	if err != nil {
		lr.E().Errorf("Failed to marshal hot data cache: %v", err)
		return fmt.Errorf("failed to marshal hot data cache: %w", err)
	}

	return cache.Set(ctx, cacheKey, string(jsonData), CoinHotDataCacheExpiration)
}

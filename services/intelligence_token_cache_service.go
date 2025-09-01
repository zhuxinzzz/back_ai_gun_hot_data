package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/model"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/pkg/model/remote"
	"back_ai_gun_data/services/remote_service"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/utils"
)

const (
	// 缓存过期时间：4天
	CacheExpiration = 4 * 24 * time.Hour
)

func ProcessIntelligenceCoinCache(data *model.MessageData) error {
	intelligenceID := data.ID

	// 从消息中提取币信息
	coins, err := extractCoinsFromMessage(data)
	if err != nil {
		lr.E().Errorf("Failed to extract coins from message: %v", err)
		return fmt.Errorf("failed to extract coins from message: %w", err)
	}

	if err := updateIntelligenceCoinCache(intelligenceID, coins); err != nil {
		lr.E().Errorf("Failed to update intelligence coin cache: %v", err)
		return fmt.Errorf("failed to update intelligence coin cache: %w", err)
	}

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
			CreatedAt: dto.CustomTime{Time: time.Now()},
			UpdatedAt: dto.CustomTime{Time: time.Now()},
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
			CreatedAt: dto.CustomTime{Time: time.Now()},
			UpdatedAt: dto.CustomTime{Time: time.Now()},
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

func updateIntelligenceCoinCache(intelligenceID string, newCoins []dto.IntelligenceCoinCache) error {
	// 获取现有缓存
	existingCoins, err := getIntelligenceCoinCache(intelligenceID)
	if err != nil {
		lr.E().Errorf("Failed to get existing cache for intelligence %s: %v", intelligenceID, err)
		// 如果获取失败，使用空数组
		existingCoins = []dto.IntelligenceCoinCache{}
	}

	// 调用GMGN服务更新市场信息
	if err := updateMarketInfoFromGMGN(newCoins); err != nil {
		lr.E().Errorf("Failed to update market info from GMGN for intelligence %s: %v", intelligenceID, err)
		// 即使GMGN调用失败，也继续更新缓存，只是市场信息可能不是最新的
	}

	// 处理币热数据流程（更新admin服务缓存、排序、打标签、进入热数据缓存）
	if err := ProcessCoinHotData(intelligenceID, newCoins); err != nil {
		lr.E().Errorf("Failed to process coin hot data for intelligence %s: %v", intelligenceID, err)
		// 热数据处理失败不影响主缓存更新流程
	}

	// 创建现有币的映射
	existingCoinsMap := make(map[string]dto.IntelligenceCoinCache)
	for _, coin := range existingCoins {
		existingCoinsMap[coin.Name] = coin
	}

	// 添加或更新新币信息
	for _, newCoin := range newCoins {
		if existingCoin, exists := existingCoinsMap[newCoin.Name]; exists {
			// 更新现有币信息，保留原有的市场数据
			newCoin.Stats = existingCoin.Stats
			newCoin.UpdatedAt = dto.CustomTime{Time: time.Now()}
		}
		existingCoinsMap[newCoin.Name] = newCoin
	}

	// 转换回切片
	var updatedCoins []dto.IntelligenceCoinCache
	for _, coin := range existingCoinsMap {
		updatedCoins = append(updatedCoins, coin)
	}

	// 更新缓存
	return setIntelligenceCoinCache(intelligenceID, updatedCoins)
}

func getIntelligenceCoinCache(intelligenceID string) ([]dto.IntelligenceCoinCache, error) {
	ctx := context.Background()
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID

	cacheData, err := cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var data []dto.IntelligenceCoinCache
	if err := json.Unmarshal([]byte(cacheData), &data); err != nil {
		lr.E().Errorf("Failed to unmarshal cache data: %v", err)
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return data, nil
}

// setIntelligenceCoinCache 设置情报-币缓存
func setIntelligenceCoinCache(intelligenceID string, data []dto.IntelligenceCoinCache) error {
	ctx := context.Background()
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID

	jsonData, err := json.Marshal(data)
	if err != nil {
		lr.E().Errorf("Failed to marshal cache data: %v", err)
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	return cache.Set(ctx, cacheKey, string(jsonData), CacheExpiration)
}

// GetIntelligenceCoins 获取情报关联的币信息
func GetIntelligenceCoins(intelligenceID string) ([]dto.IntelligenceCoinCache, error) {
	data, err := getIntelligenceCoinCache(intelligenceID)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DeleteIntelligenceCoinCache 删除情报-币缓存
func DeleteIntelligenceCoinCache(intelligenceID string) error {
	ctx := context.Background()
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID
	return cache.Del(ctx, cacheKey)
}

// updateMarketInfoFromGMGN 从GMGN服务更新市场信息
func updateMarketInfoFromGMGN(coins []dto.IntelligenceCoinCache) error {
	// 收集所有需要查询的币名称
	var coinNames []string
	var validCoins []dto.IntelligenceCoinCache

	for _, coin := range coins {
		if coin.Name != "" {
			coinNames = append(coinNames, coin.Name)
			validCoins = append(validCoins, coin)
		}
	}

	if len(coinNames) == 0 {
		return nil
	}

	// 将名称列表转换为逗号分隔的字符串
	namesStr := strings.Join(coinNames, ",")

	// 批量调用GMGN服务查询市场信息
	tokens, err := remote_service.QueryTokensByName(namesStr, "")
	if err != nil {
		lr.E().Errorf("Failed to query GMGN for coins: %v", err)
		return err
	}

	// 创建token映射，用于快速查找
	tokenMap := make(map[string]remote.GmGnToken)
	for _, token := range tokens {
		tokenMap[token.Name] = token
	}

	// 更新市场信息
	for _, coin := range validCoins {
		if token, exists := tokenMap[coin.Name]; exists {
			// 找到对应的币，更新其索引
			for j, originalCoin := range coins {
				if originalCoin.Name == coin.Name {
					coins[j].Stats.CurrentPriceUSD = token.PriceUSD
					coins[j].Stats.CurrentMarketCap = token.MarketCap
					coins[j].UpdatedAt = dto.CustomTime{Time: time.Now()}

					break
				}
			}
		}
	}

	return nil
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

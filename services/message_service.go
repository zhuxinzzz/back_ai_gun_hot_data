package services

import (
	"back_ai_gun_data/pkg/dao"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/pkg/model/dto_cache"
	"back_ai_gun_data/pkg/model/remote"
	"back_ai_gun_data/services/remote_service"
	"back_ai_gun_data/utils"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ProcessMessageData(ctx context.Context, data *model.MessageData) error {
	entities := analyzeEntities(data)

	if err := UpdateTokenMarketData(ctx, data.ID); err != nil {
		lr.E().Errorf("Failed to update admin market data: %v", err)
	}

	if err := processRankingAndHotData(ctx, data, entities); err != nil {
		lr.E().Errorf("Coin ranking failed: %v", err)
	}

	return nil
}

// 使用remote包中的SupportedChains

var top3 = 3

func processRankingAndHotData(ctx context.Context, data *model.MessageData, entities map[string]interface{}) error {
	cacheTokens, err := ReadTokenCache(ctx, data.ID)
	if err != nil {
		lr.E().Errorf("Failed to read intelligence token cache: %v", err)
		return fmt.Errorf("failed to read intelligence token cache: %w", err)
	}
	if len(cacheTokens) == 0 {
		lr.I().Infof("No cacheTokens found in cache for intelligence %s", data.ID)
		return nil
	}

	// 直接用已有的token名组装调用remote
	var searchNames []string
	for _, t := range cacheTokens {
		if t.Name != "" {
			searchNames = append(searchNames, t.Name)
		}
	}

	combined := make([]dto_cache.IntelligenceTokenCache, 0, len(cacheTokens)+len(searchNames))
	combined = append(combined, cacheTokens...) // 旧币放在前面以便稳定排序时保序

	// 2.3 为所有币名称批量查询GMGN数据，发现新币并补齐市值
	if len(searchNames) > 0 {
		// 批量查询，limit 按名称数目放大
		namesStr := strings.Join(searchNames, ",")
		limit := len(searchNames) * 3
		if limit < 10 {
			limit = 10
		}

		remoteTokens, qErr := remote_service.QueryTokensByNameWithLimit(ctx, namesStr, "", limit)
		if qErr == nil {
			searchResultsByName := make(map[string][]remote.GmGnToken)
			for _, t := range remoteTokens {
				if t.IsSupportedChain() {
					searchResultsByName[t.Name] = append(searchResultsByName[t.Name], t)
				}
			}

			// 从remote搜索结果中发现新币（不在缓存中的币）
			existingNames := make(map[string]struct{}, len(cacheTokens))
			for _, t := range cacheTokens {
				if t.Name != "" {
					existingNames[t.Name] = struct{}{}
				}
			}

			var newTokenNames []string
			for name := range searchResultsByName {
				if _, ok := existingNames[name]; !ok {
					newTokenNames = append(newTokenNames, name)
				}
			}

			for _, name := range newTokenNames {
				now := time.Now()
				newToken := dto_cache.IntelligenceTokenCache{
					ID:        utils.GenerateUUIDV7(),
					Name:      name,
					Standard:  stringPtr("ERC20"),
					Stats:     dto_cache.CoinMarketStats{},
					Chain:     dto_cache.ChainInfo{},
					CreatedAt: dto_cache.CustomTime{Time: now},
					UpdatedAt: dto_cache.CustomTime{Time: now},
				}

				// 为每个name查找所有可能的网络结果，优先选择市值最高的
				var bestToken *remote.GmGnToken
				var bestMarketCap float64

				tokens, exists := searchResultsByName[name]
				if !exists {
					combined = append(combined, newToken)
					continue
				}

				for _, token := range tokens {
					marketCap, _ := strconv.ParseFloat(token.MarketCap, 64)
					if bestToken == nil || marketCap > bestMarketCap {
						bestToken = &token
						bestMarketCap = marketCap
					}
				}

				if bestToken == nil {
					combined = append(combined, newToken)
					continue
				}

				// 填充最佳token的数据
				newToken.Stats.CurrentPriceUSD = bestToken.PriceUSD
				newToken.Stats.CurrentMarketCap = bestToken.MarketCap
				newToken.Stats.WarningPriceUSD = bestToken.PriceUSD
				newToken.Stats.WarningMarketCap = bestToken.MarketCap
				newToken.ContractAddress = bestToken.Address
				newToken.Chain.Slug = strings.ToLower(bestToken.Network)

				combined = append(combined, newToken)
			}
		} else {
			lr.E().Errorf("Batch GMGN query failed for new cacheTokens: %v", qErr)
		}
	}

	// 3. 调用admin服务进行稳定排序，返回排序后的切片
	rankedCoins, err := remote_service.CallAdminRanking(combined)
	if err != nil {
		lr.E().Errorf("Failed to call admin ranking service: %v", err)
		return fmt.Errorf("failed to call admin ranking service: %w", err)
	}

	hasNewTokenInTop3 := false
	if len(rankedCoins) < top3 {
		top3 = len(rankedCoins)
	}
	oldTokenKeys := make(map[string]dto_cache.IntelligenceTokenCache, len(cacheTokens))
	for _, t := range cacheTokens {
		oldTokenKeys[t.GetUniqueKey()] = t
	}
	for i := 0; i < top3; i++ {
		if _, exists := oldTokenKeys[rankedCoins[i].GetUniqueKey()]; !exists {
			hasNewTokenInTop3 = true
			break
		}
	}

	if hasNewTokenInTop3 {
		// 构建最终缓存：按照admin排序结果，保留旧的全部 + 新的前三
		finalCache := make([]dto_cache.IntelligenceTokenCache, 0, len(rankedCoins))

		// 遍历排序后的结果，按顺序添加
		for _, token := range rankedCoins {
			if _, exists := oldTokenKeys[token.GetUniqueKey()]; exists {
				finalCache = append(finalCache, token)
			} else {
				if len(finalCache) < top3 {
					finalCache = append(finalCache, token)
				}
			}
		}

		if err := writeTokenCache(ctx, data.ID, finalCache); err != nil {
			lr.E().Errorf("Failed to write updated token cache: %v", err)
			// 缓存写入失败不影响后续流程，继续处理热点数据
		}
	}

	if err := ProcessCoinHotData(data.ID, rankedCoins); err != nil {
		lr.E().Errorf("Failed to process token hot data: %v", err)
		return fmt.Errorf("failed to process token hot data: %w", err)
	}

	lr.I().Infof("Successfully processed token ranking, cache update and hot data for intelligence %s", data.ID)
	return nil
}

func analyzeEntities(data *model.MessageData) map[string]interface{} {
	entities := data.Data.EntitiesExtract.Entities

	result := map[string]interface{}{
		"tokens":   entities.Tokens,
		"projects": entities.Projects,
		"persons":  entities.Persons,
		"accounts": entities.Accounts,
	}

	return result
}

func createOrGetEntityFromETL(tokenName string) (*dto.Entity, error) {
	// 检查是否已存在实体
	existingEntity, err := dao.GetEntityBySlugAndType(tokenName, "token")
	if err != nil {
		lr.E().Errorf("Failed to get entity by slug and type: %v", err)
		return nil, err
	}

	if existingEntity != nil {
		return existingEntity, nil
	}

	entity := &dto.Entity{
		Name:   tokenName,
		Slug:   tokenName,
		Type:   "token",
		Source: stringPtr("etl"),
	}

	//if err := createEntity(entity); err != nil {
	//	lr.E().Errorf("Failed to create entity: %v", err)
	//	return nil, err
	//}

	return entity, nil
}

// createIntelligenceEntityRelation 创建情报实体关联
func createIntelligenceEntityRelation(intelligenceID, entityID string) error {
	relation := &dto.EntityIntelligence{
		IntelligenceID: intelligenceID,
		EntityID:       entityID,
		Type:           stringPtr("token"),
	}

	return dao.CreateEntityIntelligence(relation)
}

// triggerCoinDataSearch 触发币数据搜索
func triggerCoinDataSearch(tokenName, entityID, intelligenceID string) error {
	// 直接调用币热数据服务的搜索功能
	//return SearchCoinDataDirectly(tokenName, entityID, intelligenceID)
	return nil
}

// getTokenNameSet 从消息数据中提取token名集合的辅助函数
// 从消息的实体提取结果中获取所有token名称，用于市场数据处理和缓存维护
func getTokenNameSet(entities map[string]interface{}) map[string]bool {
	tokenNameSet := make(map[string]bool)

	// 从实体中提取tokens
	if tokens, ok := entities["tokens"].([]string); ok {
		for _, tokenName := range tokens {
			if tokenName != "" {
				tokenNameSet[tokenName] = true
			}
		}
	}

	lr.I().Infof("Extracted %d unique token names from message entities", len(tokenNameSet))
	return tokenNameSet
}

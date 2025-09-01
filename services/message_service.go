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

func processRankingAndHotData(ctx context.Context, data *model.MessageData, entities map[string]interface{}) error {
	tokens, err := ReadTokenCache(ctx, data.ID)
	if err != nil {
		lr.E().Errorf("Failed to read intelligence token cache: %v", err)
		return fmt.Errorf("failed to read intelligence token cache: %w", err)
	}
	if len(tokens) == 0 {
		lr.I().Infof("No tokens found in cache for intelligence %s", data.ID)
		return nil
	}

	existingNames := make(map[string]struct{}, len(tokens))
	for _, c := range tokens {
		if c.Name != "" {
			existingNames[c.Name] = struct{}{}
		}
	}

	newNameSet := getTokenNameSet(entities)
	var missingNames []string
	for name := range newNameSet {
		if _, ok := existingNames[name]; !ok {
			missingNames = append(missingNames, name)
		}
	}

	combined := make([]dto_cache.IntelligenceTokenCache, 0, len(tokens)+len(missingNames))
	combined = append(combined, tokens...) // 旧币放在前面以便稳定排序时保序

	// 2.3 为缺失的名称批量查询GMGN数据，尽量补齐市值
	if len(missingNames) > 0 {
		// 批量查询，limit 按名称数目放大
		namesStr := strings.Join(missingNames, ",")
		limit := len(missingNames) * 3
		if limit < 10 {
			limit = 10
		}
		gmgnTokens, qErr := remote_service.QueryTokensByNameWithLimit(ctx, namesStr, "", limit)
		if qErr != nil {
			lr.E().Errorf("Batch GMGN query failed for new tokens: %v", qErr)
			gmgnTokens = nil
		}
		gmgnByName := make(map[string]remote.GmGnToken)
		for _, t := range gmgnTokens {
			gmgnByName[t.Name] = t
		}

		// 2.4 生成新增token占位并尽量补齐市值，确保有稳定的唯一ID
		for _, name := range missingNames {
			now := time.Now()
			newToken := dto_cache.IntelligenceTokenCache{
				ID: utils.GenerateUUIDV7(),
				//EntityID: utils.GenerateUUIDV7(),
				Name:      name,
				Symbol:    generateSymbol(name),
				Standard:  stringPtr("ERC20"),
				Decimals:  18,
				Stats:     dto_cache.CoinMarketStats{},
				Chain:     dto_cache.ChainInfo{},
				CreatedAt: dto_cache.CustomTime{Time: now},
				UpdatedAt: dto_cache.CustomTime{Time: now},
			}

			if info, ok := gmgnByName[name]; ok {
				newToken.Stats.CurrentPriceUSD = info.PriceUSD
				newToken.Stats.CurrentMarketCap = info.MarketCap
				// 直接使用真实数据作为预警值，确保新token公平参与排序
				newToken.Stats.WarningPriceUSD = info.PriceUSD
				newToken.Stats.WarningMarketCap = info.MarketCap
				if newToken.ContractAddress == "" {
					newToken.ContractAddress = info.Address
				}
				// 如果需要，也可根据 info.Network 映射链信息，这里保持默认不强制设置
			}

			combined = append(combined, newToken)
		}
	}

	// 3. 调用admin服务进行稳定排序，返回排序后的切片
	rankedCoins, err := remote_service.CallAdminRanking(combined)
	if err != nil {
		lr.E().Errorf("Failed to call admin ranking service: %v", err)
		return fmt.Errorf("failed to call admin ranking service: %w", err)
	}

	// 4. 只有当新token进入前三名时，才将合并后的完整数据写回情报缓存
	// 检查前三名中是否有新增token
	hasNewTokenInTop3 := false
	top3 := 3
	if len(rankedCoins) < top3 {
		top3 = len(rankedCoins)
	}

	// 构建旧缓存名称集合，用于快速判断
	oldTokenNames := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		oldTokenNames[t.Name] = struct{}{}
	}

	// 检查前三名中是否有新增token
	for i := 0; i < top3; i++ {
		if _, exists := oldTokenNames[rankedCoins[i].Name]; !exists {
			hasNewTokenInTop3 = true
			break
		}
	}

	// 只有新token进入前三名时才更新缓存
	if hasNewTokenInTop3 {
		// 构建最终缓存：按照admin排序结果，保留旧的全部 + 新的前三
		finalCache := make([]dto_cache.IntelligenceTokenCache, 0, len(rankedCoins))

		// 遍历排序后的结果，按顺序添加
		for _, token := range rankedCoins {
			if _, exists := oldTokenNames[token.Name]; exists {
				// 旧token，直接添加（保持排序后的顺序）
				finalCache = append(finalCache, token)
			} else {
				// 新token，只有在前三名才添加
				if len(finalCache) < 3 {
					finalCache = append(finalCache, token)
				}
			}
		}

		if err := writeTokenCache(ctx, data.ID, finalCache); err != nil {
			lr.E().Errorf("Failed to write updated token cache: %v", err)
			// 缓存写入失败不影响后续流程，继续处理热点数据
		}
		lr.I().Infof("Updated cache with admin ranking order: %d tokens for %s", len(finalCache), data.ID)
	}

	// 5. 处理热点数据（仅追加新的前三名）
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

// 辅助函数

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

package services

import (
	"back_ai_gun_data/pkg/dao"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/services/remote_service"
	"context"
	"fmt"
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
	// 阶段二：市场数据enriquecimiento与首次排名

	// 1. 从缓存读取最新的币数据（上层已经更新过市场信息）
	coins, err := ReadTokenCache(ctx, data.ID)
	if err != nil {
		lr.E().Errorf("Failed to read intelligence token cache: %v", err)
		return fmt.Errorf("failed to read intelligence token cache: %w", err)
	}

	if len(coins) == 0 {
		lr.I().Infof("No coins found in cache for intelligence %s", data.ID)
		return nil
	}

	// 2. 调用admin服务排序接口，返回排序后的切片
	rankedCoins, err := remote_service.CallAdminRanking(coins)
	if err != nil {
		lr.E().Errorf("Failed to call admin ranking service: %v", err)
		return fmt.Errorf("failed to call admin ranking service: %w", err)
	}

	// 3. 处理热点数据（仅追加新的前三名）
	if err := ProcessCoinHotData(data.ID, rankedCoins); err != nil {
		lr.E().Errorf("Failed to process token hot data: %v", err)
		return fmt.Errorf("failed to process token hot data: %w", err)
	}

	lr.I().Infof("Successfully processed token ranking and hot data for intelligence %s", data.ID)
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

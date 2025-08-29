package services

import (
	"back_ai_gun_data/pkg/dao"
	"back_ai_gun_data/pkg/model"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/services/remote_service"
	"fmt"
	"strings"
	"time"

	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/utils"
)

/*
阶段二：市场数据 enriquecimiento 与首次排名
获取市场数据：币服务 开始维护 admin服务 缓存中的市场信息。它会调用外部数据源 gmgn，获取关键的市场指标，主要是 current_price_usd（当前美元价格）和 current_market_cap（当前市值）。
调用排序：币服务 在更新了市场数据后，会调用 admin服务 提供的排序接口，对相关的币进行排名。
打热点标签：admin服务 完成排序后，币服务 获取排序结果。对于排名前三的币，币服务 会将其 is_show 字段标记为 true。这个标签的含义是“该币种曾经达到过市场排名前三”，是一个重要的荣誉标记。
进入热数据缓存：任何一个被打上 is_show 标签的币，其完整的币信息都会被 币服务 存入一个专门的“币热数据”缓存中。这个缓存汇集了所有曾经进入过前三名的币。
*/

func ProcessMessageData(data *model.MessageData) error {
	entities := analyzeEntities(data)

	if err := maintainAdminMarketData(data, entities); err != nil {
		lr.E().Errorf("Admin market data failed: %v", err)
	}

	if err := processCoinRankingAndHotData(data, entities); err != nil {
		lr.E().Errorf("Coin ranking failed: %v", err)
	}

	return nil
}

func maintainAdminMarketData(data *model.MessageData, entities map[string]interface{}) error {
	// 直接调用admin服务更新市场信息
	// admin服务会从缓存读取数据，然后使用GMGN更新市场信息
	if err := remote_service.UpdateAdminMarketData(data.ID); err != nil {
		return fmt.Errorf("admin update failed: %w", err)
	}

	return nil
}

func processCoinRankingAndHotData(data *model.MessageData, entities map[string]interface{}) error {
	// TODO: 实现排序和热数据处理逻辑
	// 这里需要从缓存读取数据，调用排序服务，然后更新热数据缓存
	return nil
}

// 分析实体信息
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

func processCoinHotDataFromETL(data *model.MessageData, entities map[string]interface{}) error {
	// 使用辅助函数获取token名集合
	tokenNameSet := getTokenNameSet(entities)
	if len(tokenNameSet) == 0 {
		return nil // 没有token实体，不需要处理
	}

	// 转换为列表用于处理
	tokenNames := getTokenNameList(entities)

	// 将token名称列表转换为逗号分隔的字符串
	tokenNamesStr := joinTokenNames(tokenNames)
	if tokenNamesStr != "" {
		// 这里可以调用远程服务，例如：
		// remote_service.QueryTokensByName(tokenNamesStr)
		_, err := remote_service.QueryTokensByName(tokenNamesStr, "")
		if err != nil {
			lr.E().Error(err)
			return err
		}
	}

	return nil
}

func processTokenFromETL(tokenName string, data *model.MessageData) error {
	// 创建或获取情报
	intelligence, err := createOrGetIntelligenceFromETL(tokenName, data)
	if err != nil {
		lr.E().Errorf("Failed to create/get intelligence for token %s: %v", tokenName, err)
		return fmt.Errorf("failed to create/get intelligence for token %s: %w", tokenName, err)
	}

	// 创建或获取实体
	entity, err := createOrGetEntityFromETL(tokenName)
	if err != nil {
		lr.E().Errorf("Failed to create/get entity for token %s: %v", tokenName, err)
		return fmt.Errorf("failed to create/get entity for token %s: %w", tokenName, err)
	}

	// 创建情报实体关联
	if err := createIntelligenceEntityRelation(intelligence.ID, entity.ID); err != nil {
		lr.E().Errorf("Failed to create intelligence-entity relation: %v", err)
		return fmt.Errorf("failed to create intelligence-entity relation: %w", err)
	}

	// 触发币数据搜索和更新
	if err := triggerCoinDataSearch(tokenName, entity.ID, intelligence.ID); err != nil {
		lr.E().Errorf("Failed to trigger coin data search: %v", err)
		return fmt.Errorf("failed to trigger coin data search: %w", err)
	}

	return nil
}

func createOrGetIntelligenceFromETL(tokenName string, data *model.MessageData) (*dto.Intelligence, error) {
	existingIntelligence, err := findIntelligenceByToken(tokenName)
	if err != nil {
		return nil, err
	}

	if existingIntelligence != nil {
		// 更新现有情报
		content := data.Data.Content
		existingIntelligence.Content = &content
		existingIntelligence.SourceURL = data.Data.SourceURL
		existingIntelligence.UpdatedAt = time.Now()

		if err := updateIntelligence(existingIntelligence); err != nil {
			return nil, err
		}

		return existingIntelligence, nil
	}

	// 创建新情报
	title := fmt.Sprintf("Token Intelligence: %s", tokenName)
	content := data.Data.Content
	intelligence := &dto.Intelligence{
		Title:       &title,
		Content:     &content,
		SourceURL:   data.Data.SourceURL,
		Type:        "token_intelligence",
		PublishedAt: time.Unix(data.Data.PublishedAt, 0),
		SourceID:    utils.GenerateUUIDV7(), // 生成一个source_id
		IsValuable:  false,
		Score:       0.0,
	}

	if err := dao.CreateIntelligence(intelligence); err != nil {
		lr.E().Errorf("Failed to create intelligence: %v", err)
		return nil, err
	}

	return intelligence, nil
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

	if err := createEntity(entity); err != nil {
		lr.E().Errorf("Failed to create entity: %v", err)
		return nil, err
	}

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

// getTokenNameList 从消息数据中提取token名称列表的辅助函数
// 返回所有token名称的字符串切片，用于需要列表形式的场景
func getTokenNameList(entities map[string]interface{}) []string {
	tokenNameSet := getTokenNameSet(entities)

	tokenNames := make([]string, 0, len(tokenNameSet))
	for tokenName := range tokenNameSet {
		tokenNames = append(tokenNames, tokenName)
	}

	return tokenNames
}

// joinTokenNames 将token名称列表转换为逗号分隔的字符串
// 输入: []string{"name1", "name2", "name3"}
// 输出: "name1,name2,name3"
func joinTokenNames(tokenNames []string) string {
	if len(tokenNames) == 0 {
		return ""
	}

	// 使用strings.Join是最高效的方式
	return strings.Join(tokenNames, ",")
}

func findIntelligenceByToken(tokenName string) (*dto.Intelligence, error) {
	// TODO: 实现根据token名称查找情报的逻辑
	return nil, nil
}

func updateIntelligence(intelligence *dto.Intelligence) error {
	// TODO: 实现更新情报的逻辑
	return nil
}

func createEntity(entity *dto.Entity) error {
	// TODO: 实现创建实体的逻辑
	return nil
}

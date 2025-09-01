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
	if err := remote_service.UpdateAdminMarketData(nil, data.ID); err != nil {
		return fmt.Errorf("admin update failed: %w", err)
	}

	return nil
}

func processCoinRankingAndHotData(data *model.MessageData, entities map[string]interface{}) error {
	// 阶段二：市场数据enriquecimiento与首次排名

	// 1. 更新admin服务缓存中的市场信息
	if err := remote_service.UpdateAdminMarketData(nil, data.ID); err != nil {
		lr.E().Errorf("Failed to update admin market data: %v", err)
		// 继续执行，不中断流程
	}

	// 2. 从缓存读取最新的币数据
	coins, err := remote_service.ReadIntelligenceCoinCacheFromRedis(data.ID)
	if err != nil {
		lr.E().Errorf("Failed to read intelligence coin cache: %v", err)
		return fmt.Errorf("failed to read intelligence coin cache: %w", err)
	}

	if len(coins) == 0 {
		lr.I().Infof("No coins found in cache for intelligence %s", data.ID)
		return nil
	}

	// 3. 调用admin服务排序接口
	rankingResponse, err := remote_service.CallAdminRankingService(coins)
	if err != nil {
		lr.E().Errorf("Failed to call admin ranking service: %v", err)
		return fmt.Errorf("failed to call admin ranking service: %w", err)
	}

	// 4. 处理热点标签并进入热数据缓存
	if err := ProcessCoinHotData(data.ID, rankingResponse.Data); err != nil {
		lr.E().Errorf("Failed to process coin hot data: %v", err)
		return fmt.Errorf("failed to process coin hot data: %w", err)
	}

	lr.I().Infof("Successfully processed coin ranking and hot data for intelligence %s", data.ID)
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

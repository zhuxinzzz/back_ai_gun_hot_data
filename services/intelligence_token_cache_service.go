package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/dao"
	"back_ai_gun_data/pkg/model"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/pkg/model/dto_cache"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"back_ai_gun_data/pkg/lr"
)

const (
	// 缓存过期时间：4天
	CacheExpiration = 4 * 24 * time.Hour
)

func getIntelligenceCoinCache(intelligenceID string) ([]dto_cache.IntelligenceToken, error) {
	ctx := context.Background()
	cacheKey := IntelligenceCoinCacheKeyPrefix + intelligenceID

	cacheData, err := cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var data []dto_cache.IntelligenceToken
	if err := json.Unmarshal([]byte(cacheData), &data); err != nil {
		lr.E().Errorf("Failed to unmarshal cache data: %v", err)
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return data, nil
}

func GetIntelligenceCoins(intelligenceID string) ([]dto_cache.IntelligenceToken, error) {
	data, err := getIntelligenceCoinCache(intelligenceID)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SyncShowedTokensToIntelligence(intelligenceID string) error {
	cacheTokens, err := GetIntelligenceCoins(intelligenceID)
	if err != nil {
		lr.E().Errorf("Failed to get intelligence coins for sync: %v", err)
		return err
	}

	showedTokens := make([]dto.ShowedToken, 0, len(cacheTokens))
	for _, ct := range cacheTokens {
		showedTokens = append(showedTokens, ct.ToShowedToken())
	}

	if err := dao.UpdateIntelligenceShowedTokens(intelligenceID, showedTokens); err != nil {
		lr.E().Errorf("Failed to update intelligence showed tokens: %v", err)
		return err
	}

	return nil
}

func stringPtr(s string) *string {
	return &s
}

func ProcessIntelligenceData(ctx context.Context, data *model.IntelligenceMessage) error {
	//lr.I().Infof("Processing intelligence data: %+v", data)

	// TODO: 实现情报排序数据的处理逻辑
	// 1. 解析情报数据
	// 2. 进行排序计算
	// 3. 更新数据库
	// 4. 缓存结果

	err := ProcessMessageData(ctx, data)
	if err != nil {
		lr.E().Error(err)
		return err
	}

	return nil
}

func ProcessETLEntityData(ctx context.Context, data *model.ETLEntityMessage) error {
	//lr.I().Infof("Processing ETL entity data: %+v", data)

	// TODO: 实现ETL实体数据的处理逻辑
	// 1. 解析AI分析出的实体数据
	// 2. 同步币信息
	// 3. 更新实体关联
	// 4. 触发相关业务逻辑
	err := TriggerMarketDataUpdate(ctx, data.ID)
	if err != nil {
		lr.E().Error(err)
		return err
	}
	return nil
}

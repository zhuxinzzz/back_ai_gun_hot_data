package services

import (
	"back_ai_gun_data/pkg/dao"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/pkg/model/dto_cache"
	"back_ai_gun_data/pkg/model/remote"
	"back_ai_gun_data/services/remote_service"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ProcessMessageData(ctx context.Context, data *model.MessageData) error {
	entities := analyzeEntities(data)

	// 触发市场数据更新
	err := UpdateMarketData(ctx, data.ID)
	if err != nil {
		lr.E().Error(err)
	}

	if err := processRankingAndHotData(ctx, data, entities); err != nil {
		lr.E().Error(err)
		return err
	}

	return nil
}

var top3 = 3

const (
	detectionInterval = 30 * time.Second // 30秒检测间隔
	maxDetections     = 10               // 最多检测10次
)

func processRankingAndHotData(ctx context.Context, data *model.MessageData, entities map[string]interface{}) error {
	time.Sleep(detectionInterval)

	cacheTokens, err := ReadTokenCache(ctx, data.ID)
	if err != nil {
		lr.E().Error(err)
		return err
	}
	if len(cacheTokens) == 0 {
		lr.I().Infof("No cacheTokens found in cache for intelligence %s", data.ID)
		return nil
	}

	// 构建搜索名称列表
	var searchNames []string
	for _, t := range cacheTokens {
		if t.Name != "" {
			searchNames = append(searchNames, t.Name)
		}
	}

	// 执行10次定时检测
	for detectionCount := 0; detectionCount < maxDetections; detectionCount++ {
		select {
		case <-ctx.Done():
			lr.I().Infof("Context cancelled, stopping detection for intelligence %s", data.ID)
			return nil
		default:
			// 执行一次检测和处理
			if err := executeDetectionAndProcessing(ctx, data.ID, searchNames, cacheTokens); err != nil {
				lr.E().Errorf("Detection %d failed: %v", detectionCount+1, err)
				// 继续下一次检测，不中断流程
			}

			detectionCount++
			lr.I().Infof("Detection %d/%d completed for intelligence %s", detectionCount, maxDetections, data.ID)

			// 如果不是最后一次检测，等待下次检测
			if detectionCount < maxDetections {
				time.Sleep(detectionInterval)
			}
		}
	}

	lr.I().Infof("Completed %d detections for intelligence %s", maxDetections, data.ID)
	return nil
}

// executeDetectionAndProcessing 执行一次检测和处理
func executeDetectionAndProcessing(ctx context.Context, intelligenceID string, searchNames []string, cacheTokens []dto_cache.IntelligenceTokenCache) error {
	combined := make([]dto_cache.IntelligenceTokenCache, 0, len(cacheTokens)+len(searchNames))
	combined = append(combined, cacheTokens...) // 旧币放在前面以便稳定排序时保序

	// 2.3 为所有币名称批量查询GMGN数据，发现新币并补齐市值
	if len(searchNames) > 0 {
		remoteTokens, qErr := queryTokensByName(ctx, searchNames)
		if qErr == nil {
			searchResultsByName := make(map[string][]remote.GmGnToken)
			for _, t := range remoteTokens {
				if t.IsSupportedChain() {
					searchResultsByName[t.Name] = append(searchResultsByName[t.Name], t)
				}
			}

			// 从remote搜索结果中发现新币（不在缓存中的币）
			var newTokens []dto_cache.IntelligenceTokenCache
			for _, tokens := range searchResultsByName {
				for _, token := range tokens {
					// 检查是否已存在相同的币种
					isNewToken := true
					for _, existingToken := range cacheTokens {
						if existingToken.IsSameToken(token) {
							isNewToken = false
							break
						}
					}

					if isNewToken {
						newToken := toIntelligenceTokenCache(token, "ERC20")
						newTokens = append(newTokens, newToken)
					}
				}
			}

			for i := range newTokens {
				tokens, exists := searchResultsByName[newTokens[i].Name]
				if !exists {
					combined = append(combined, newTokens[i])
					continue
				}

				// 选择市值最高的token
				var bestToken *remote.GmGnToken
				var bestMarketCap float64
				for _, token := range tokens {
					if token.Name == newTokens[i].Name &&
						strings.EqualFold(token.Address, newTokens[i].ContractAddress) &&
						strings.EqualFold(token.Network, newTokens[i].Chain.Slug) {
						marketCap, _ := strconv.ParseFloat(token.MarketCap, 64)
						if bestToken == nil || marketCap > bestMarketCap {
							bestToken = &token
							bestMarketCap = marketCap
						}
					}
				}

				if bestToken != nil {
					// 填充市场数据
					newTokens[i].Stats.CurrentPriceUSD = bestToken.PriceUSD
					newTokens[i].Stats.CurrentMarketCap = bestToken.MarketCap
					newTokens[i].Stats.WarningPriceUSD = bestToken.PriceUSD
					newTokens[i].Stats.WarningMarketCap = bestToken.MarketCap
				}

				combined = append(combined, newTokens[i])
			}
		} else {
			lr.E().Error(qErr)
			return qErr
		}
	}

	// 3. 调用admin服务进行稳定排序，返回排序后的切片
	rankedCoins, err := remote_service.CallAdminRanking(combined)
	if err != nil {
		lr.E().Error(err)
		return err
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

		if err := writeTokenCache(ctx, intelligenceID, finalCache); err != nil {
			lr.E().Error(err)
			// 缓存写入失败不影响后续流程，继续处理热点数据
		}
	}

	if err := ProcessCoinHotData(intelligenceID, rankedCoins); err != nil {
		lr.E().Error(err)
		return err
	}

	return nil
}

// startTokenDetectionTask 启动定时检测新币任务
func startTokenDetectionTask(ctx context.Context, intelligenceID string, searchNames []string) {
	const (
		detectionInterval = 30 * time.Second // 30秒检测间隔
		maxDetections     = 10               // 最多检测10次
		maxRetries        = 9                // 最多重试9次
	)

	// 使用context控制任务生命周期
	taskCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 防止重复启动的标记
	taskKey := fmt.Sprintf("token_detection_%s", intelligenceID)
	if isTaskRunning(taskKey) {
		lr.I().Infof("Token detection task already running for intelligence %s", intelligenceID)
		return
	}
	setTaskRunning(taskKey, true)
	defer setTaskRunning(taskKey, false)

	detectionCount := 0
	retryCount := 0

	for detectionCount < maxDetections {
		select {
		case <-ctx.Done():
			lr.I().Infof("Context cancelled, stopping token detection for intelligence %s", intelligenceID)
			return
		case <-taskCtx.Done():
			lr.I().Infof("Task cancelled, stopping token detection for intelligence %s", intelligenceID)
			return
		default:
			// 执行检测
			if hasNewTokens := detectNewTokens(ctx, intelligenceID, searchNames); hasNewTokens {
				lr.I().Infof("Found new tokens in detection %d for intelligence %s", detectionCount+1, intelligenceID)
				// 发现新币继续检测，不停止任务
			} else {
				lr.I().Infof("No new tokens found in detection %d for intelligence %s", detectionCount+1, intelligenceID)
			}

			detectionCount++
			lr.I().Infof("Detection %d/%d completed for intelligence %s", detectionCount, maxDetections, intelligenceID)

			// 如果10次检测完成，开始重试流程
			if detectionCount >= maxDetections && retryCount < maxRetries {
				retryCount++
				detectionCount = 0 // 重置检测计数
				lr.I().Infof("Starting retry %d/%d for intelligence %s", retryCount, maxRetries, intelligenceID)
			}

			// 等待下次检测
			time.Sleep(detectionInterval)
		}
	}

	lr.I().Infof("Token detection task completed for intelligence %s after %d detections and %d retries", intelligenceID, detectionCount, retryCount)
}

// queryTokensByName 查询GMGN数据
func queryTokensByName(ctx context.Context, searchNames []string) ([]remote.GmGnToken, error) {
	if len(searchNames) == 0 {
		return nil, nil
	}

	// 查询GMGN
	namesStr := strings.Join(searchNames, ",")
	limit := len(searchNames) * 3
	if limit < 10 {
		limit = 10
	}

	remoteTokens, err := remote_service.QueryTokensByNameWithLimit(ctx, namesStr, "", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query GMGN: %w", err)
	}

	return remoteTokens, nil
}

// detectNewTokens 检测是否有新币
func detectNewTokens(ctx context.Context, intelligenceID string, searchNames []string) bool {
	// 获取当前缓存
	cacheTokens, err := ReadTokenCache(ctx, intelligenceID)
	if err != nil {
		lr.E().Errorf("Failed to read cache for detection: %v", err)
		return false
	}

	// 不需要构建唯一键集合，直接使用IsSameToken方法比较

	// 查询GMGN
	remoteTokens, err := queryTokensByName(ctx, searchNames)
	if err != nil {
		lr.E().Errorf("Failed to query GMGN for detection: %v", err)
		return false
	}

	// 检查是否有新币
	for _, token := range remoteTokens {
		if token.IsSupportedChain() {
			// 检查是否已存在相同的币种
			for _, existingToken := range cacheTokens {
				if existingToken.IsSameToken(token) {
					goto nextToken // 找到相同币种，检查下一个
				}
			}
			// 没有找到相同币种，说明是新币
			lr.I().Infof("Found new token %s on %s chain during detection for intelligence %s", token.Name, token.Network, intelligenceID)
			return true
		nextToken:
		}
	}

	return false
}

// 简单的任务状态管理
var (
	taskRunningMap = make(map[string]bool)
	taskMutex      sync.RWMutex
)

// toIntelligenceTokenCache 将GmGnToken转换为IntelligenceTokenCache
// 不足的字段通过参数传入
func toIntelligenceTokenCache(token remote.GmGnToken, standard string) dto_cache.IntelligenceTokenCache {
	return dto_cache.IntelligenceTokenCache{
		Name:            token.Name,
		Symbol:          token.Symbol,
		Standard:        &standard,
		Decimals:        token.Decimals,
		ContractAddress: token.Address,
		Logo:            token.Logo,
		Chain: dto_cache.ChainInfo{
			Slug: strings.ToLower(token.Network),
		},
		CreatedAt: dto_cache.CustomTime{Time: time.Now()},
		UpdatedAt: dto_cache.CustomTime{Time: time.Now()},
	}
}

func isTaskRunning(key string) bool {
	taskMutex.RLock()
	defer taskMutex.RUnlock()
	return taskRunningMap[key]
}

func setTaskRunning(key string, running bool) {
	taskMutex.Lock()
	defer taskMutex.Unlock()
	taskRunningMap[key] = running
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
		lr.E().Error(err)
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
	//	lr.E().Error(err)
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

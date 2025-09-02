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

var top3 = 3

const (
	detectionInterval = 30 * time.Second // 30秒检测间隔
	maxDetections     = 10               // 最多检测10次
)

func ProcessMessageData(ctx context.Context, data *model.MessageData) error {
	entities := analyzeEntities(data)

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

	searchNames := make([]string, 0, len(cacheTokens))
	searchAddresses := make([]string, 0, len(cacheTokens))
	for _, token := range cacheTokens {
		if token.Name != "" {
			searchNames = append(searchNames, token.Name)
		}
		if token.ContractAddress != "" {
			searchAddresses = append(searchAddresses, token.ContractAddress)
		}
	}

	dtoTokens, err := dao.GetProjectChainDataByNamesAndAddresses(searchNames, searchAddresses)
	if err != nil {
		lr.E().Error(err)
		return err
	}

	convertedTokens := convertProjectChainDataToCacheTokens(dtoTokens)
	cacheTokens = append(cacheTokens, convertedTokens...)

	for detectionCount := 0; detectionCount < maxDetections; detectionCount++ {

		select {
		case <-ctx.Done():
			lr.I().Infof("Context cancelled, stopping detection for intelligence %s", data.ID)
			return nil
		default:
			if err := executeDetectionAndProcessing(ctx, data.ID, searchNames, cacheTokens); err != nil {
				lr.E().Errorf("Detection %d failed: %v", detectionCount+1, err)
				// 继续下一次检测，不中断流程
			}

			detectionCount++

			// 如果不是最后一次检测，等待下次检测
			if detectionCount < maxDetections {
				time.Sleep(detectionInterval)
			}
		}
	}

	return nil
}

func executeDetectionAndProcessing(ctx context.Context, intelligenceID string, searchNames []string, cacheTokens []dto_cache.IntelligenceToken) error {
	combined := make([]dto_cache.IntelligenceToken, 0, len(cacheTokens)+len(searchNames))
	combined = append(combined, cacheTokens...) // 旧币放在前面以便稳定排序时保序

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
			var newTokens []dto_cache.IntelligenceToken
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

	rankedTokens, err := remote_service.CallAdminRankingWithCache(intelligenceID, combined)
	if err != nil {
		lr.E().Error(err)
		return err
	}

	hasNewTokenInTop3 := false
	if len(rankedTokens) < top3 {
		top3 = len(rankedTokens)
	}
	oldTokenKeys := make(map[string]dto_cache.IntelligenceToken, len(cacheTokens))
	for _, t := range cacheTokens {
		oldTokenKeys[t.GetUniqueKey()] = t
	}
	for i := 0; i < top3; i++ {
		if _, exists := oldTokenKeys[rankedTokens[i].GetUniqueKey()]; !exists {
			hasNewTokenInTop3 = true
			break
		}
	}

	if hasNewTokenInTop3 {
		finalCache := make([]dto_cache.IntelligenceToken, 0, len(rankedTokens))

		// 遍历排序后的结果，按顺序添加
		for _, token := range rankedTokens {
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

	if err := SyncShowedTokensToIntelligence(intelligenceID); err != nil {
		lr.E().Error(err)
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
	namesStr := strings.Join(searchNames, ",")
	limit := len(searchNames) * 3
	if limit < 10 {
		limit = 10
	}

	remoteTokens, err := remote_service.QueryTokensByNameWithLimit(ctx, namesStr, "", limit)
	if err != nil {
		lr.E().Error(err)
		return nil, err
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
func toIntelligenceTokenCache(token remote.GmGnToken, standard string) dto_cache.IntelligenceToken {
	return dto_cache.IntelligenceToken{
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

// convertProjectChainDataToCacheTokens 将 ProjectChainData 转换为 IntelligenceToken
func convertProjectChainDataToCacheTokens(dtoTokens []*dto.ProjectChainData) []dto_cache.IntelligenceToken {
	cacheTokens := make([]dto_cache.IntelligenceToken, 0, len(dtoTokens))

	for _, dtoToken := range dtoTokens {
		if dtoToken == nil {
			continue
		}
		// 跳过没有名称或合约地址的记录
		if dtoToken.Name == nil || *dtoToken.Name == "" || dtoToken.ContractAddress == "" {
			continue
		}

		// 获取链信息
		chainInfo := dto_cache.ChainInfo{
			ID:   "",
			Name: *dtoToken.Name,
			Slug: "",
			Logo: "",
		}

		// 如果有 ChainID，可以查询链信息（这里简化处理）
		if dtoToken.ChainID != nil {
			// TODO: 可以根据需要查询链信息
			chainInfo.ID = *dtoToken.ChainID
		}

		// 创建本地变量并设置默认值
		var currentPriceUSD string
		var currentMarketCap string

		// 安全赋值，判断数据源是否为空
		if dtoToken.Price24Hours != nil {
			currentPriceUSD = fmt.Sprintf("%.8f", *dtoToken.Price24Hours)
		}
		if dtoToken.MarketCap24Hours != nil {
			currentMarketCap = fmt.Sprintf("%.2f", *dtoToken.MarketCap24Hours)
		}

		// 构建市场统计信息
		stats := dto_cache.CoinMarketStats{
			WarningPriceUSD:     currentPriceUSD,
			WarningMarketCap:    currentMarketCap,
			CurrentPriceUSD:     currentPriceUSD,
			CurrentMarketCap:    currentMarketCap,
			HighestIncreaseRate: "",
		}

		var symbol string
		var decimals int
		var logo string
		var entityID string
		if dtoToken.Symbol != nil {
			symbol = *dtoToken.Symbol
		}
		if dtoToken.Decimals != nil {
			decimals = *dtoToken.Decimals
		}
		if dtoToken.Logo != nil {
			logo = *dtoToken.Logo
		}
		if dtoToken.EntityID != nil {
			entityID = *dtoToken.EntityID
		}

		// 构建 IntelligenceToken
		cacheToken := dto_cache.IntelligenceToken{
			ID:              dtoToken.ID,
			EntityID:        entityID,
			Name:            *dtoToken.Name,
			Symbol:          symbol,
			Standard:        dtoToken.Standard,
			Decimals:        decimals,
			ContractAddress: dtoToken.ContractAddress,
			Logo:            logo,
			Stats:           stats,
			Chain:           chainInfo,
			CreatedAt:       dto_cache.CustomTime{Time: dtoToken.CreatedAt},
			UpdatedAt:       dto_cache.CustomTime{Time: dtoToken.UpdatedAt},
		}

		cacheTokens = append(cacheTokens, cacheToken)
	}

	return cacheTokens
}

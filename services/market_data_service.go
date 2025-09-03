package services

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto_cache"
	"back_ai_gun_data/services/remote_service"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func UpdateMarketData(ctx context.Context, intelligenceID string) error {
	cacheData, err := ReadTokenCache(ctx, intelligenceID)
	if err != nil {
		lr.E().Errorf("Failed to read intelligence token cache: %v", err)
		return fmt.Errorf("failed to read intelligence token cache: %w", err)
	}
	if len(cacheData) == 0 {
		lr.E().Errorf("No cacheTokens found in cache for intelligence %s", intelligenceID)
		return nil
	}

	// 批量获取GMGN数据
	updatedCount := 0

	// 收集所有需要查询的币种信息
	var queryParams []struct {
		index int
		token dto_cache.IntelligenceToken
	}

	for i, token := range cacheData {
		if token.Name != "" {
			queryParams = append(queryParams, struct {
				index int
				token dto_cache.IntelligenceToken
			}{i, token})
		}
	}

	if len(queryParams) == 0 {
		lr.E().Errorf("No valid cacheTokens to query for intelligence %s", intelligenceID)
		return nil
	}

	// 收集所有币的名称
	var coinNames []string
	for _, param := range queryParams {
		coinNames = append(coinNames, param.token.Name)
	}

	// 批量查询所有币
	namesStr := strings.Join(coinNames, ",")
	limit := len(coinNames) * 2
	if limit < 10 {
		limit = 10 // 最少10个
	}

	// 直接调用上游接口，不指定链
	remoteTokens, err := remote_service.QueryTokensByNameWithLimit(ctx, namesStr, "", limit)
	if err != nil {
		lr.E().Errorf("Failed to batch query GMGN: %v", err)
		return fmt.Errorf("failed to batch query GMGN: %w", err)
	}

	// 更新每个币的市场信息
	for _, param := range queryParams {
		cacheTokenIns := param.token
		index := param.index

		// 使用提取的匹配方法
		matchedToken := cacheTokenIns.FindMatchingToken(remoteTokens)

		// 更新市场信息
		if matchedToken != nil {
			cacheData[index].Stats.CurrentPriceUSD = matchedToken.PriceUSD
			cacheData[index].Stats.CurrentMarketCap = matchedToken.MarketCap
			cacheData[index].UpdatedAt.Time = time.Now()

			// 计算预警涨幅：当前市值 ÷ 预警市值
			if matchedToken.MarketCap != "0" && cacheTokenIns.Stats.WarningMarketCap != "0" {
				currentMarketCap, err1 := strconv.ParseFloat(matchedToken.MarketCap, 64)
				warningMarketCap, err2 := strconv.ParseFloat(cacheTokenIns.Stats.WarningMarketCap, 64)

				if err1 == nil && err2 == nil && warningMarketCap > 0 {
					currentIncreaseRate := currentMarketCap / warningMarketCap

					// 获取历史最高涨幅
					highestIncreaseRate, err3 := strconv.ParseFloat(cacheTokenIns.Stats.HighestIncreaseRate, 64)
					if err3 != nil {
						highestIncreaseRate = 0
					}

					// 更新最高涨幅（取较大值）
					if currentIncreaseRate > highestIncreaseRate {
						cacheData[index].Stats.HighestIncreaseRate = fmt.Sprintf("%.6f", currentIncreaseRate)
					}
				}
			}

			updatedCount++
		} else {
			//lr.E().Errorf("No GMGN data found for token: %s", cacheTokenIns.Name)
		}
	}

	// 将更新后的数据写回缓存
	if err := writeTokenCache(ctx, intelligenceID, cacheData); err != nil {
		lr.E().Errorf("Failed to write intelligence token cache: %v", err)
		return fmt.Errorf("failed to write intelligence token cache: %w", err)
	}

	// 同步showed_tokens到intelligence表
	//if err := SyncShowedTokensToIntelligence(intelligenceID); err != nil {
	//	lr.E().Errorf("Failed to sync showed tokens to intelligence: %v", err)
	//	// 不返回错误，因为缓存更新已经成功
	//}

	return nil
}

func TriggerMarketDataUpdate(ctx context.Context, intelligenceID string) (err error) {
	if err := UpdateMarketData(ctx, intelligenceID); err != nil {
		lr.E().Errorf("Failed to update market data for intelligence %s: %v", intelligenceID, err)
		return err
	}
	return nil
}

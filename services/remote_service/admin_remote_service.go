package remote_service

import (
	"back_ai_gun_data/pkg/model/dto_cache"
	"sort"
	"strconv"
	"time"
)

// CallAdminRanking 对代币集合进行排序并返回排序后的结果（按市值降序）
func CallAdminRanking(coins []dto_cache.IntelligenceTokenCache) ([]dto_cache.IntelligenceTokenCache, error) {
	// 空输入直接返回空输出
	if len(coins) == 0 {
		return []dto_cache.IntelligenceTokenCache{}, nil
	}

	// 复制一份数据进行排序，避免修改原数据
	rankedCoins := make([]dto_cache.IntelligenceTokenCache, len(coins))
	copy(rankedCoins, coins)

	// 按市值降序排序
	// 优先使用current_market_cap，如果为0则使用warning_market_cap
	sort.Slice(rankedCoins, func(i, j int) bool {
		// 获取i的市值
		iMarketCap := rankedCoins[i].Stats.CurrentMarketCap
		if iMarketCap == "0" || iMarketCap == "" {
			iMarketCap = rankedCoins[i].Stats.WarningMarketCap
		}

		// 获取j的市值
		jMarketCap := rankedCoins[j].Stats.CurrentMarketCap
		if jMarketCap == "0" || jMarketCap == "" {
			jMarketCap = rankedCoins[j].Stats.WarningMarketCap
		}

		// 转换为float进行比较
		iVal, err1 := strconv.ParseFloat(iMarketCap, 64)
		jVal, err2 := strconv.ParseFloat(jMarketCap, 64)

		// 如果解析失败，按字符串比较
		if err1 != nil || err2 != nil {
			return iMarketCap > jMarketCap
		}

		// 按数值降序排列
		return iVal > jVal
	})

	// 模拟网络延迟
	time.Sleep(10 * time.Millisecond)

	return rankedCoins, nil
}

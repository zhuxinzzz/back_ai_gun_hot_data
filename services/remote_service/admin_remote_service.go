package remote_service

import (
	"back_ai_gun_data/pkg/model/dto"
	"sort"
	"strconv"
	"time"
)

func CallAdminRanking(coins []dto.IntelligenceCoinCache) (*dto.AdminRankingResponse, error) {
	// 模拟admin服务排序接口
	// 接口：POST /api/admin/ranking
	// Body: 代币集合
	// 返回：排序后的代币集合（按市值降序排列）

	if len(coins) == 0 {
		return &dto.AdminRankingResponse{
			Code:    0,
			Message: "success",
			Data:    []dto.IntelligenceCoinCache{},
		}, nil
	}

	// 复制一份数据进行排序，避免修改原数据
	rankedCoins := make([]dto.IntelligenceCoinCache, len(coins))
	copy(rankedCoins, coins)

	// 按市值降序排序（模拟admin服务的排序逻辑）
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

	// 返回排序后的数据
	response := &dto.AdminRankingResponse{
		Code:    0,
		Message: "success",
		Data:    rankedCoins,
	}

	return response, nil
}

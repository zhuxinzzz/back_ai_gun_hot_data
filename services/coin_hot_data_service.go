package services

import (
	"back_ai_gun_data/pkg/model/dto_cache"
)

// ProcessCoinHotData 处理币热数据流程（单层实现）
// 输入的 coins 应为已按排名排序的列表
// 逻辑：仅当当前前三中出现“新成员”时，按顺序追加到缓存；已存在的历史成员保留
func ProcessCoinHotData(intelligenceID string, coins []dto_cache.IntelligenceTokenCache) error {
	// 只关心当前排名的前三名
	topN := 3
	if len(coins) < topN {
		topN = len(coins)
	}
	if topN == 0 {
		return nil
	}

	// 读取现有热数据缓存（为一个全局列表）
	existing, err := ReadTokenCache(nil, intelligenceID)
	if err != nil || existing == nil {
		existing = []dto_cache.IntelligenceTokenCache{}
	}

	// 构建已存在ID集合，避免重复
	existingIDs := make(map[string]struct{}, len(existing))
	for _, c := range existing {
		existingIDs[c.ID] = struct{}{}
	}

	// 按顺序检查前三，如果是新成员则按顺序追加
	initialLen := len(existing)
	for i := 0; i < topN; i++ {
		rc := coins[i]
		if _, ok := existingIDs[rc.ID]; ok {
			continue
		}
		existing = append(existing, rc)
	}

	// 若没有新增则不写回
	if len(existing) == initialLen {
		return nil
	}

	return writeTokenCache(nil, intelligenceID, existing)
}

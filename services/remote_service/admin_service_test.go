package remote_service

import (
	"back_ai_gun_data/pkg/model/dto"
	"testing"
)

func TestCallAdminRankingService(t *testing.T) {
	// 创建测试数据
	coins := []dto.IntelligenceCoinCache{
		{
			ID:   "1",
			Name: "Bitcoin",
			Stats: dto.CoinMarketStats{
				CurrentMarketCap: "50000000000", // 500亿
				WarningMarketCap: "10000000000", // 100亿
			},
		},
		{
			ID:   "2",
			Name: "Ethereum",
			Stats: dto.CoinMarketStats{
				CurrentMarketCap: "30000000000", // 300亿
				WarningMarketCap: "8000000000",  // 80亿
			},
		},
		{
			ID:   "3",
			Name: "Cardano",
			Stats: dto.CoinMarketStats{
				CurrentMarketCap: "0",          // 0，会使用warning_market_cap
				WarningMarketCap: "5000000000", // 50亿
			},
		},
		{
			ID:   "4",
			Name: "Polkadot",
			Stats: dto.CoinMarketStats{
				CurrentMarketCap: "8000000000", // 80亿
				WarningMarketCap: "2000000000", // 20亿
			},
		},
	}

	// 调用排序服务
	response, err := CallAdminRankingService(coins)
	if err != nil {
		t.Fatalf("CallAdminRankingService failed: %v", err)
	}

	// 验证响应
	if response.Code != 0 {
		t.Errorf("Expected code 0, got %d", response.Code)
	}

	if len(response.Data) != len(coins) {
		t.Errorf("Expected %d coins, got %d", len(coins), len(response.Data))
	}

	// 验证排序结果（按市值降序）
	expectedOrder := []string{"Bitcoin", "Ethereum", "Polkadot", "Cardano"}
	for i, coin := range response.Data {
		if coin.Name != expectedOrder[i] {
			t.Errorf("Expected %s at position %d, got %s", expectedOrder[i], i, coin.Name)
		}
	}

	t.Logf("Sorting test passed! Ranked order: %v", expectedOrder)
}

func TestCallAdminRankingServiceEmpty(t *testing.T) {
	// 测试空数组
	response, err := CallAdminRankingService([]dto.IntelligenceCoinCache{})
	if err != nil {
		t.Fatalf("CallAdminRankingService failed: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("Expected code 0, got %d", response.Code)
	}

	if len(response.Data) != 0 {
		t.Errorf("Expected empty data, got %d items", len(response.Data))
	}

	t.Logf("Empty array test passed!")
}

package remote_service

import (
	"back_ai_gun_data/pkg/model/dto_cache"
	"testing"
)

func TestCallAdminRankingServiceEmpty(t *testing.T) {
	// 测试空数组
	response, err := CallAdminRanking([]dto_cache.IntelligenceTokenCache{})
	if err != nil {
		t.Fatalf("CallAdminRanking failed: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("Expected code 0, got %d", response.Code)
	}

	if len(response.Data) != 0 {
		t.Errorf("Expected empty data, got %d items", len(response.Data))
	}

	t.Logf("Empty array test passed!")
}

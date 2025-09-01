package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto_cache"
	"back_ai_gun_data/pkg/model/remote"
	"encoding/json"
	"fmt"
)

const (
	AdminRankingURL = "/api/v1/sort"
)

func getAdminHost() string {
	return "http://192.168.4.64:8001"
}

func CallAdminRanking(coins []dto_cache.IntelligenceTokenCache) ([]dto_cache.IntelligenceTokenCache, error) {
	requestData := map[string]interface{}{
		"tokens": coins,
	}

	resp, err := Cli().R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestData).
		Post(getAdminHost() + AdminRankingURL)
	if err != nil {
		lr.E().Error(err)
		return nil, err
	}
	if resp.StatusCode() != 200 {
		lr.E().Errorf("Admin ranking API returned status %d: %s", resp.StatusCode(), resp.String())
		return nil, fmt.Errorf("admin ranking API error: status %d", resp.StatusCode())
	}

	var response remote.AdminRankingResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		lr.E().Errorf("Failed to unmarshal admin ranking response: %v", err)
		return nil, err
	}
	if response.Code != 0 {
		lr.E().Errorf("Admin ranking API error: %s", response.Message)
		return nil, fmt.Errorf("admin ranking API error: %s", response.Message)
	}

	// 将interface{}转换为具体类型
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		lr.E().Errorf("Failed to marshal response data: %v", err)
		return nil, err
	}

	var tokens []dto_cache.IntelligenceTokenCache
	if err := json.Unmarshal(dataBytes, &tokens); err != nil {
		lr.E().Errorf("Failed to unmarshal tokens: %v", err)
		return nil, err
	}

	return tokens, nil
}

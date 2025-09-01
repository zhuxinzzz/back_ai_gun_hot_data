package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/pkg/model/dto_cache"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/gjson"
)

const (
	AdminRankingURL = "/api/v1/sort/"
)

func getAdminHost() string {
	return "http://192.168.4.64:8001"
}

// ConvertCacheToSortRequest 将 []dto_cache.IntelligenceTokenCache 转换为 dto.SortRequest
func ConvertCacheToSortRequest(intelligenceID string, cacheTokens []dto_cache.IntelligenceTokenCache) dto.SortRequest {
	var tokenList []dto.TokenReq
	for _, cache := range cacheTokens {
		token := dto.TokenReq{
			ID:              cache.ID,
			EntityID:        cache.EntityID,
			Name:            cache.Name,
			Symbol:          cache.Symbol,
			Standard:        cache.Standard,
			Decimals:        cache.Decimals,
			ContractAddress: cache.ContractAddress,
			Logo:            cache.Logo,
			Stats: dto.CoinMarketStats{
				WarningPriceUSD:     cache.Stats.WarningPriceUSD,
				WarningMarketCap:    cache.Stats.WarningMarketCap,
				CurrentPriceUSD:     cache.Stats.CurrentPriceUSD,
				CurrentMarketCap:    cache.Stats.CurrentMarketCap,
				HighestIncreaseRate: cache.Stats.HighestIncreaseRate,
			},
			Chain: dto.ChainInfo{
				ID:   cache.Chain.ID,
				Name: cache.Chain.Name,
				Slug: cache.Chain.Slug,
				Logo: cache.Chain.Logo,
			},
			CreatedAt: dto.CustomTime{Time: cache.CreatedAt.Time},
			UpdatedAt: dto.CustomTime{Time: cache.UpdatedAt.Time},
		}
		tokenList = append(tokenList, token)
	}

	return dto.SortRequest{
		IntelligenceID:      intelligenceID,
		IntelligenceHotData: []dto.TokenReq{},
		TokenList:           tokenList,
	}
}

// ConvertSortResponseToCache 将 []dto.IntelligenceTokenCacheResp 转换为 []dto_cache.IntelligenceTokenCache
func ConvertSortResponseToCache(dtoTokens []dto.IntelligenceTokenCacheResp) []dto_cache.IntelligenceTokenCache {
	var result []dto_cache.IntelligenceTokenCache
	for _, dtoToken := range dtoTokens {
		// 解析时间字符串
		var createdAt, updatedAt dto_cache.CustomTime
		if dtoToken.CreatedAt != "" {
			if t, err := time.Parse("2006-01-02T15:04:05.000", dtoToken.CreatedAt); err == nil {
				createdAt = dto_cache.CustomTime{Time: t}
			}
		}
		if dtoToken.UpdatedAt != "" {
			if t, err := time.Parse("2006-01-02T15:04:05.000", dtoToken.UpdatedAt); err == nil {
				updatedAt = dto_cache.CustomTime{Time: t}
			}
		}

		cache := dto_cache.IntelligenceTokenCache{
			ID:              dtoToken.ID,
			EntityID:        dtoToken.EntityID,
			Name:            dtoToken.Name,
			Symbol:          dtoToken.Symbol,
			Standard:        dtoToken.Standard,
			Decimals:        dtoToken.Decimals,
			ContractAddress: dtoToken.ContractAddress,
			Logo:            dtoToken.Logo,
			Stats: dto_cache.CoinMarketStats{
				WarningPriceUSD:     dtoToken.Stats.WarningPriceUSD,
				WarningMarketCap:    dtoToken.Stats.WarningMarketCap,
				CurrentPriceUSD:     dtoToken.Stats.CurrentPriceUSD,
				CurrentMarketCap:    dtoToken.Stats.CurrentMarketCap,
				HighestIncreaseRate: dtoToken.Stats.HighestIncreaseRate,
			},
			Chain: dto_cache.ChainInfo{
				ID:   dtoToken.Chain.ID,
				Name: dtoToken.Chain.Name,
				Slug: dtoToken.Chain.Slug,
				Logo: dtoToken.Chain.Logo,
			},
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		result = append(result, cache)
	}
	return result
}

func CallAdminRanking(req dto.SortRequest) ([]dto.IntelligenceTokenCacheResp, error) {
	resp, err := Cli().R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(getAdminHost() + AdminRankingURL)
	if err != nil {
		lr.E().Error(err)
		return nil, err
	}

	if resp.StatusCode() != 200 {
		lr.E().Errorf("Admin ranking API returned status %d: %s", resp.StatusCode(), resp.String())
		return nil, fmt.Errorf("admin ranking API error: status %d", resp.StatusCode())
	}

	var tokenList []dto.IntelligenceTokenCacheResp
	dataStr := gjson.Get(resp.String(), "data").Raw
	if err := json.Unmarshal([]byte(dataStr), &tokenList); err != nil {
		lr.E().Errorf("Failed to unmarshal data array: %v", err)
		return nil, err
	}

	return tokenList, nil
}

func CallAdminRankingWithCache(intelligenceID string, cacheTokens []dto_cache.IntelligenceTokenCache) ([]dto_cache.IntelligenceTokenCache, error) {
	req := ConvertCacheToSortRequest(intelligenceID, cacheTokens)

	resp, err := CallAdminRanking(req)
	if err != nil {
		lr.E().Error(err)
		return nil, err
	}

	return ConvertSortResponseToCache(resp), nil
}

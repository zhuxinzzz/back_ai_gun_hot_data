package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/pkg/model/dto_cache"
	"encoding/json"
	"fmt"
	"time"

	"back_ai_gun_data/pkg/model/remote"

	"github.com/tidwall/gjson"
)

const (
	AdminRankingURL = "/api/v1/sort/"
)

func getAdminHost() string {
	return "http://192.168.4.64:8001"
}

func callAdminRanking(req dto.RankReq) ([]dto.IntelligenceTokenRankResp, error) {
	urlIns := getAdminHost() + AdminRankingURL
	resp, err := Cli().R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(urlIns)
	if err != nil {
		lr.E().Error(err)
		return nil, err
	}

	if resp.StatusCode() != 200 {
		lr.E().Errorf("Admin ranking API returned status %d: %s", resp.StatusCode(), resp.String())
		return nil, fmt.Errorf("admin ranking API error: status %d", resp.StatusCode())
	}

	var tokenList []dto.IntelligenceTokenRankResp
	dataStr := gjson.Get(resp.String(), "data").Raw
	if err := json.Unmarshal([]byte(dataStr), &tokenList); err != nil {
		lr.E().Errorf("Failed to unmarshal data array: %v", err)
		return nil, err
	}

	return tokenList, nil
}

func convertCacheToTokenReq(cacheTokens []dto_cache.IntelligenceToken) []dto.OldTokenReq {
	tokenList := make([]dto.OldTokenReq, 0, len(cacheTokens))
	for _, cache := range cacheTokens {
		tokenList = append(tokenList, cache.ToOldTokenReq())
	}
	return tokenList
}

func CallAdminRankingWithGmGnTokens(intelligenceID string, oldTokens []dto_cache.IntelligenceToken, newTokens []remote.GmGnToken) ([]dto_cache.IntelligenceToken, error) {
	req := dto.RankReq{
		IntelligenceID:      intelligenceID,
		IntelligenceHotData: convertCacheToTokenReq(oldTokens)[:1],         // todo test
		TokenList:           convertGmGnTokensToNewTokenReq(newTokens)[:1], // todo test
	}

	resp, err := callAdminRanking(req)
	if err != nil {
		lr.E().Error(err)
		return nil, err
	}

	return ConvertSortResponseToCache(resp), nil
}

func convertGmGnTokensToNewTokenReq(gmgnTokens []remote.GmGnToken) []dto.NewTokenReq {
	tokenList := make([]dto.NewTokenReq, 0, len(gmgnTokens))
	for _, token := range gmgnTokens {
		tokenList = append(tokenList, token.ToNewTokenReq())
	}
	return tokenList
}

func ConvertSortResponseToCache(dtoTokens []dto.IntelligenceTokenRankResp) []dto_cache.IntelligenceToken {
	result := make([]dto_cache.IntelligenceToken, 0, len(dtoTokens))
	for _, dtoToken := range dtoTokens {
		// 判断是否为外部API数据结构
		isExternalAPI := dtoToken.Network != "" || dtoToken.PriceUSD != "" || dtoToken.Volume24h != ""

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

		var cache dto_cache.IntelligenceToken

		if isExternalAPI {
			// 外部API数据结构处理
			cache = dto_cache.IntelligenceToken{
				ID:              "", // 外部API没有ID
				EntityID:        "", // 外部API没有EntityID
				Name:            dtoToken.Name,
				Symbol:          dtoToken.Symbol,
				Standard:        nil, // 外部API没有Standard
				Decimals:        dtoToken.Decimals,
				ContractAddress: dtoToken.ContractAddress,
				Logo:            dtoToken.Logo,
				Stats: dto_cache.CoinMarketStats{
					WarningPriceUSD:     "0", // 外部API没有预警价格
					WarningMarketCap:    "0", // 外部API没有预警市值
					CurrentPriceUSD:     dtoToken.PriceUSD,
					CurrentMarketCap:    dtoToken.CurrentMarketCap,
					HighestIncreaseRate: "0", // 外部API没有涨幅信息
				},
				Chain: dto_cache.ChainInfo{
					ID:        "",                  // 外部API没有链ID
					NetworkID: "",                  // 外部API没有NetworkID
					Name:      dtoToken.Chain.Name, // 从Chain字段获取
					Symbol:    "",                  // 外部API没有Symbol
					Slug:      dtoToken.Network,    // 使用Network字段作为Slug
					Logo:      "",                  // 外部API没有链Logo
				},
				CreatedAt: dto_cache.CustomTime{}, // 外部API没有创建时间
				UpdatedAt: dto_cache.CustomTime{}, // 外部API没有更新时间
			}
		} else {
			// 内部数据结构处理（原有逻辑）
			cache = dto_cache.IntelligenceToken{
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
					ID:        dtoToken.Chain.ID,
					NetworkID: dtoToken.Chain.NetworkID,
					Name:      dtoToken.Chain.Name,
					Symbol:    dtoToken.Chain.Symbol,
					Slug:      dtoToken.Chain.Slug,
					Logo:      dtoToken.Chain.Logo,
				},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}
		}

		result = append(result, cache)
	}
	return result
}

package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/remote"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

const (
	queryTokensURL   = "/api/v1/ai/tokens"
	tokenSecurityURL = "/api/v1/ai/tokens/%s/security"
)

func QueryTokens(ctx context.Context, params remote.TokenQueryParams) (*remote.TokenQueryResponse, error) {
	queryParams := url.Values{}
	if params.Q != "" {
		queryParams.Set("q", params.Q)
	}
	if params.Chain != "" {
		queryParams.Set("chain", params.Chain)
	}
	if params.Limit > 0 {
		queryParams.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Fuzzy >= 0 {
		queryParams.Set("fuzzy", strconv.Itoa(params.Fuzzy))
	}

	apiURL := GetHost() + queryTokensURL
	// 手动构建查询字符串，保留空格
	if len(queryParams) > 0 {
		pairs := make([]string, 0, len(queryParams))
		for key, values := range queryParams {
			for _, value := range values {
				// 使用 url.QueryEscape 保留空格为 %20
				escapedKey := url.QueryEscape(key)
				escapedValue := url.QueryEscape(value)
				pairs = append(pairs, escapedKey+"="+escapedValue)
			}
		}
		apiURL += "?" + strings.Join(pairs, "&")
	}

	// 发送请求
	resp, err := Cli().R().
		SetContext(ctx).
		Get(apiURL)

	if err != nil {
		lr.E().Error("QueryTokens failed: ", err)
		return nil, fmt.Errorf("query tokens failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("http error: %d", resp.StatusCode())
	}

	var result remote.TokenQueryResponse
	if err := jsoniter.Unmarshal(resp.Body(), &result); err != nil {
		lr.E().Error("Failed to unmarshal response: ", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查业务错误
	if result.Code != 0 {
		return nil, fmt.Errorf("business error: %d - %s", result.Code, result.Msg)
	}

	return &result, nil
}

func QueryTokensByName(name string, chain string) ([]remote.GmGnToken, error) {
	return QueryTokensByNameWithLimit(nil, name, chain, 10) // 默认10，保持向后兼容
}

// SOL→ BSC→ ETH→ Base
var chainFilter = map[string]struct{}{
	"solana":   {},
	"bsc":      {},
	"ethereum": {},
	"eth":      {},
	"base":     {},
}

func QueryTokensByNameWithLimit(ctx context.Context, name string, chain string, limit int) ([]remote.GmGnToken, error) {
	params := remote.TokenQueryParams{
		Q:     name,  // 查询关键字,全称、简称、地址,多个查询用逗号分隔
		Chain: chain, // 指定链
		Limit: limit, // 由调用者控制数量
		Fuzzy: 1,     // 是否为模糊匹配,1:是,0:否 默认值: 1
	}

	resp, err := QueryTokens(ctx, params)
	if err != nil {
		lr.E().Error("QueryTokensByNameWithLimit failed: ", err)
		return nil, err
	}

	var allTokens []remote.GmGnToken
	for _, tokens := range resp.Data {
		allTokens = append(allTokens, tokens...)
	}

	fillteredTokens := make([]remote.GmGnToken, 0, len(allTokens))
	for _, token := range allTokens {
		if _, exists := chainFilter[token.Network]; exists {
			fillteredTokens = append(fillteredTokens, token)
		}
	}

	return fillteredTokens, nil
}

func QueryTokenSecurity(address string, platform string) (*remote.TokenSecurityResp, error) {
	apiURL := GetHost() + fmt.Sprintf(tokenSecurityURL, address) + "?platform=" + platform
	resp, err := Cli().R().Get(apiURL)
	if err != nil {
		lr.E().Error("QueryTokenSecurity failed: ", err)
		return nil, err
	}

	var result remote.TokenSecurityResp
	dataStr := gjson.Get(resp.String(), "data").Raw
	if err := jsoniter.Unmarshal([]byte(dataStr), &result); err != nil {
		lr.E().Errorf("Failed to unmarshal data: %v", err)
		return nil, err
	}

	return &result, nil
}

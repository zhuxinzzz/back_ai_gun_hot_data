package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/remote"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func QueryTokens(ctx context.Context, params remote.TokenQueryParams) (*remote.TokenQueryResponse, error) {
	return QueryTokensWithContext(ctx, params)
}

func QueryTokensWithContext(ctx context.Context, params remote.TokenQueryParams) (*remote.TokenQueryResponse, error) {
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

	apiURL := GetHost() + "/api/v1/ai/tokens"
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
	resp, err := GetCli().R().
		SetContext(ctx).
		SetResult(&remote.TokenQueryResponse{}).
		Get(apiURL)

	if err != nil {
		lr.E().Error("QueryTokens failed: ", err)
		return nil, fmt.Errorf("query tokens failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("http error: %d", resp.StatusCode())
	}

	result := resp.Result().(*remote.TokenQueryResponse)

	// 检查业务错误
	if result.Code != 0 {
		return nil, fmt.Errorf("business error: %d - %s", result.Code, result.Message)
	}

	return result, nil
}

func QueryTokensByName(name string, chain string) ([]remote.GmGnToken, error) {
	return QueryTokensByNameWithLimit(nil, name, chain, 10) // 默认10，保持向后兼容
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

	return allTokens, nil
}

func QueryTokensByAddress(address string, chain string) ([]remote.GmGnToken, error) {
	params := remote.TokenQueryParams{
		Q:     address,
		Chain: chain,
		Limit: 10,
		Fuzzy: 0, // 地址查询不使用模糊匹配
	}

	resp, err := QueryTokens(nil, params)
	if err != nil {
		lr.E().Error("QueryTokensByAddress failed: ", err)
		return nil, err
	}

	// 合并所有结果
	var allTokens []remote.GmGnToken
	for _, tokens := range resp.Data {
		allTokens = append(allTokens, tokens...)
	}

	return allTokens, nil
}

package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/remote"
	"context"
	"fmt"
	"net/url"
	"strconv"
)

func QueryTokens(params remote.TokenQueryParams) (*remote.TokenQueryResponse, error) {
	return QueryTokensWithContext(context.Background(), params)
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
	if len(queryParams) > 0 {
		apiURL += "?" + queryParams.Encode()
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
	params := remote.TokenQueryParams{
		Q:     name,  // 查询关键字,全称、简称、地址,多个查询用逗号分隔
		Chain: chain, // 指定链
		Limit: 10,    // 指定数量 默认值: 10
		Fuzzy: 1,     // 是否为模糊匹配,1:是,0:否 默认值: 1
	}

	resp, err := QueryTokens(params)
	if err != nil {
		lr.E().Error("QueryTokensByName failed: ", err)
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

	resp, err := QueryTokens(params)
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

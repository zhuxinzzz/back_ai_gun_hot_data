package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/services/remote_service"
	"back_ai_gun_data/utils"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateAdminMarketData(t *testing.T) {
	lr.Init()
	cache.Init()
	remote_service.Init()

	intelligenceID := "0198f0a9-0e77-721b-99df-b94e851375d1"
	ctx := context.Background()

	cacheTokens, err := readTokenCache(ctx, intelligenceID)
	assert.NoError(t, err)

	err = UpdateTokenMarketData(ctx, intelligenceID)
	assert.NoError(t, err)

	cacheTokens2, err := readTokenCache(ctx, intelligenceID)
	assert.NoError(t, err)

	if cacheTokens != nil && cacheTokens2 != nil {
		marketCap := cacheTokens2[0].Stats.CurrentMarketCap
		newMarketCap := cacheTokens2[0].Stats.CurrentMarketCap
		assert.True(t, marketCap != newMarketCap)
	}
}

func Test_readIntelligenceCoinCacheFromRedis(t *testing.T) {
	lr.Init()
	cache.Init()
	intelligenceID := "0198f0a9-0e77-721b-99df-b94e851375d1"
	ctx := context.Background()

	res, err := readTokenCache(ctx, intelligenceID)
	assert.NoError(t, err)
	t.Log(utils.ToJson(res))
}

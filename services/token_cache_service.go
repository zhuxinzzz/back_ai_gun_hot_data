package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto_cache"
	"context"
	"encoding/json"
	"time"
)

const (
	IntelligenceCoinCacheKeyPrefix = "dogex:intelligence:latest_entities:intelligence_id:"
)

//var chainName = map[string]struct{}{
//	"Base":             {},
//	"Polygon zkEVM":    {},
//	"Fantom":           {},
//	"Blast":            {},
//	"Arbitrum":         {},
//	"Scroll":           {},
//	"Gnosis":           {},
//	"Avalanche":        {},
//	"Polygon":          {},
//	"BSC":              {},
//	"Optimism":         {},
//	"Solana":           {},
//	"zkSync":           {},
//	"Linea":            {},
//	"Moonbeam":         {},
//	"Metis":            {},
//	"Immutable zkEVM":  {},
//	"Lisk":             {},
//	"Soneium":          {},
//	"Fuse":             {},
//	"Lens":             {},
//	"Mode":             {},
//	"Gravity":          {},
//	"Cronos":           {},
//	"Abstract":         {},
//	"Taiko":            {},
//	"Boba":             {},
//	"Unichain":         {},
//	"opBNB":            {},
//	"Swellchain":       {},
//	"Aurora":           {},
//	"Sei":              {},
//	"Sonic":            {},
//	"Moonriver":        {},
//	"Corn":             {},
//	"zkfair":           {},
//	"Bitcoin":          {},
//	"Conflux":          {},
//	"Celo":             {},
//	"Sui":              {},
//	"binancecoin":      {},
//	"World Chain":      {},
//	"the-open-network": {},
//	//"Ink":                     {},
//	"BOB":                     {},
//	"Superposition":           {},
//	"kucoin-community-chain":  {},
//	"arbitrum-nova":           {},
//	"XDC":                     {},
//	"Apechain":                {},
//	"Kaia":                    {},
//	"Avalanche X-Chain":       {},
//	"Etherlink":               {},
//	"BNB Beacon Chain (BEP2)": {},
//	"Starknet":                {},
//	"Rootstock":               {},
//	"Berachain":               {},
//	"Manta Pacific":           {},
//	"Mantle":                  {},
//	"Ethereum":                {},
//	"HyperEVM":                {},
//}

func ReadTokenCache(ctx context.Context, intelligenceID string) ([]dto_cache.IntelligenceTokenCache, error) {
	key := IntelligenceCoinCacheKeyPrefix + intelligenceID

	dataStr, err := cache.Get(ctx, key)
	if err != nil {
		// 如果缓存不存在，返回空数据而不是错误
		if err.Error() == "redis: nil" {
			lr.E().Error(err)
			return []dto_cache.IntelligenceTokenCache{}, nil
		}
		return nil, err
	}

	// 直接解析为币数组
	var coins []dto_cache.IntelligenceTokenCache
	if err := json.Unmarshal([]byte(dataStr), &coins); err != nil {
		lr.E().Error(err)
		return nil, err
	}

	return coins, nil
}

func writeTokenCache(ctx context.Context, intelligenceID string, coins []dto_cache.IntelligenceTokenCache) error {
	key := IntelligenceCoinCacheKeyPrefix + intelligenceID

	dataBytes, err := json.Marshal(coins)
	if err != nil {
		lr.E().Error(err)
		return err
	}

	// 获取原TTL，如果不存在则使用默认值
	ttl, err := cache.MainRedis().TTL(ctx, key).Result()
	if err != nil {
		lr.E().Error(err)
		// 使用默认TTL
		ttl = 4 * 24 * time.Hour
	} else if ttl < 0 {
		// TTL为-1表示没有过期时间，-2表示键不存在，使用默认TTL
		ttl = 4 * 24 * time.Hour
	}

	// 写入Redis缓存，保留原TTL
	if err := cache.Set(ctx, key, string(dataBytes), ttl); err != nil {
		lr.E().Error(err)
		return err
	}

	return nil
}

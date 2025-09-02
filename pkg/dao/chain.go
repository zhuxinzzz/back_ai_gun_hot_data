package dao

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"errors"
	"time"

	"gorm.io/gorm"
)

// GetChainByCoinGeckoChainName 根据 CoinGecko 链名称获取链信息
func GetChainByCoinGeckoChainName(coinGeckoChainName string) (*dto.Chain, error) {
	var chain dto.Chain
	result := pgDB.Where("coin_gecko_chain_name = ? AND is_deleted = false", coinGeckoChainName).First(&chain)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			lr.I().Infof("No chain found for CoinGecko chain name: %s", coinGeckoChainName)
			return nil, nil
		}
		lr.E().Errorf("Failed to get chain by CoinGecko chain name: %v", result.Error)
		return nil, result.Error
	}

	return &chain, nil
}

// GetChainUUIDByCoinGeckoChainName 根据 CoinGecko 链名称获取链 UUID
func GetChainUUIDByCoinGeckoChainName(coinGeckoChainName string) (string, error) {
	chain, err := GetChainByCoinGeckoChainName(coinGeckoChainName)
	if err != nil {
		return "", err
	}

	if chain == nil {
		return "", nil
	}

	return chain.ID, nil
}

// GetChainByCoinMarketCapChainName 根据 CoinMarketCap 链名称获取链信息
func GetChainByCoinMarketCapChainName(coinMarketCapChainName string) (*dto.Chain, error) {
	var chain dto.Chain
	result := GetDB().Where("coin_market_cap_chain_name = ?", coinMarketCapChainName).First(&chain)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			lr.I().Infof("No chain found for CoinMarketCap chain name: %s", coinMarketCapChainName)
			return nil, nil
		}
		lr.E().Errorf("Failed to get chain by CoinMarketCap chain name: %v", result.Error)
		return nil, result.Error
	}

	return &chain, nil
}

// GetChainUUIDByCoinMarketCapChainName 根据 CoinMarketCap 链名称获取链 UUID
func GetChainUUIDByCoinMarketCapChainName(coinMarketCapChainName string) (string, error) {
	chain, err := GetChainByCoinMarketCapChainName(coinMarketCapChainName)
	if err != nil {
		return "", err
	}

	if chain == nil {
		return "", nil
	}

	return chain.ID, nil
}

func GetChainByCMCChainName(coinMarketCapChainName string) (*dto.Chain, error) {
	chain, err := GetChainByCoinMarketCapChainName(coinMarketCapChainName)
	if err != nil {
		return nil, err
	}

	return chain, nil
}

func CreateChain(db *gorm.DB, chain *dto.Chain) error {
	result := db.Create(chain)
	if result.Error != nil {
		lr.E().Errorf("Failed to create chain: %v", result.Error)
		return result.Error
	}

	lr.I().Infof("Successfully created chain with UUID: %s", chain.ID)
	return nil
}

func UpdateChain(db *gorm.DB, chain *dto.Chain) error {
	result := db.Save(chain)
	if result.Error != nil {
		lr.E().Errorf("Failed to update chain: %v", result.Error)
		return result.Error
	}

	lr.I().Infof("Successfully updated chain with UUID: %s", chain.ID)
	return nil
}

// UpdateChainCoinMarketCapChainName 只更新coin_market_cap_chain_name字段
func UpdateChainCoinMarketCapChainName(db *gorm.DB, chainID string, coinMarketCapChainName string) error {
	result := db.Model(&dto.Chain{}).
		Where("id = ?", chainID).
		Update("coin_market_cap_chain_name", coinMarketCapChainName)

	if result.Error != nil {
		lr.E().Errorf("Failed to update coin_market_cap_chain_name for chain %s: %v", chainID, result.Error)
		return result.Error
	}

	return nil
}

// GetChainByUUID 根据 UUID 获取链信息
func GetChainByUUID(db *gorm.DB, uuid string) (*dto.Chain, error) {
	var chain dto.Chain
	result := db.Where("uuid = ? AND deleted_at = 0", uuid).First(&chain)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			lr.I().Infof("No chain found for UUID: %s", uuid)
			return nil, nil
		}
		lr.E().Errorf("Failed to get chain by UUID: %v", result.Error)
		return nil, result.Error
	}

	return &chain, nil
}

func GetAllChains(db *gorm.DB) ([]dto.Chain, error) {
	var chains []dto.Chain
	//result := db.Where("deleted_at = 0").Find(&chains)
	result := db.Find(&chains)
	if result.Error != nil {
		lr.E().Errorf("Failed to get all chains: %v", result.Error)
		return nil, result.Error
	}

	return chains, nil
}

// DeleteChain 软删除链记录
func DeleteChain(db *gorm.DB, uuid string) error {
	result := db.Model(&dto.Chain{}).Where("id = ?", uuid).Update("deleted_at", time.Now().UnixMilli())
	if result.Error != nil {
		lr.E().Errorf("Failed to delete chain: %v", result.Error)
		return result.Error
	}

	lr.I().Infof("Successfully deleted chain with UUID: %s", uuid)
	return nil
}

// GetChainBySlug 根据slug获取链信息
func GetChainBySlug(slug string) (*dto.Chain, error) {
	var chain dto.Chain
	result := pgDB.Where("slug = ? AND is_deleted = false", slug).First(&chain)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			lr.I().Infof("No chain found for slug: %s", slug)
			return nil, nil
		}
		lr.E().Errorf("Failed to get chain by slug: %v", result.Error)
		return nil, result.Error
	}

	return &chain, nil
}

// GetChainIDBySlug 根据slug获取链ID
func GetChainIDBySlug(slug string) (string, error) {
	chain, err := GetChainBySlug(slug)
	if err != nil {
		return "", err
	}

	if chain == nil {
		return "", nil
	}

	return chain.ID, nil
}

// GetChainsByIDs 根据链ID列表批量获取链信息
func GetChainsByIDs(chainIDs []string) (map[string]*dto.Chain, error) {
	if len(chainIDs) == 0 {
		return make(map[string]*dto.Chain), nil
	}

	var chains []dto.Chain
	result := pgDB.Where("id IN ? AND is_deleted = false", chainIDs).Find(&chains)
	if result.Error != nil {
		lr.E().Errorf("Failed to get chains by IDs: %v", result.Error)
		return nil, result.Error
	}

	// 构建 ID 到 Chain 的映射
	chainMap := make(map[string]*dto.Chain)
	for i := range chains {
		chainMap[chains[i].ID] = &chains[i]
	}

	return chainMap, nil
}

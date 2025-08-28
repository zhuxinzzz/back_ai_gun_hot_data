package dao

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"

	"gorm.io/gorm"
)

func SaveRecordsToDatabase(db *gorm.DB, records []dto.Token) error {
	// 转换DTO为模型
	var tokens []dto.Token
	for _, record := range records {
		token := dto.Token{
			Name:   record.Name,
			Symbol: record.Symbol,
			//Decimals:  record.Decimals,
			//Address:   record.Address,
			//Chain:     record.Chain,
			//Volume24h: record.Volume24h,
			//MarketCap: record.MarketCap,
			LogoURL: record.LogoURL,
		}
		tokens = append(tokens, token)
	}

	// 批量插入
	result := db.Create(&tokens)
	if result.Error != nil {
		lr.E().Error(result.Error)
		return result.Error
	}

	lr.I().Infof("Successfully saved %d records to database", result.RowsAffected)
	return nil
}

// 根据链和地址查询代币
func GetTokenByChainAndAddress(db *gorm.DB, chain, address string) (*dto.Token, error) {
	var token dto.Token
	result := db.Where("chain = ? AND address = ?", chain, address).First(&token)
	if result.Error != nil {
		return nil, result.Error
	}
	return &token, nil
}

// 根据符号查询代币
func GetTokensBySymbol(symbol string) ([]dto.Token, error) {
	var tokens []dto.Token
	result := GetDB().Where("symbol = ?", symbol).Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

// 根据链查询代币
func GetTokensByChain(db *gorm.DB, chain string) ([]dto.Token, error) {
	var tokens []dto.Token
	result := db.Where("chain = ?", chain).Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

// 分页查询代币
func GetTokensWithPagination(db *gorm.DB, offset, limit int) ([]dto.Token, error) {
	var tokens []dto.Token
	result := db.Offset(offset).Limit(limit).Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

// 更新代币信息
func UpdateToken(db *gorm.DB, token *dto.Token) error {
	result := db.Save(token)
	if result.Error != nil {
		lr.E().Error(result.Error)
		return result.Error
	}
	return nil
}

// 删除代币
func DeleteToken(db *gorm.DB, id uint) error {
	result := db.Delete(&dto.Token{}, id)
	if result.Error != nil {
		lr.E().Error(result.Error)
		return result.Error
	}
	return nil
}

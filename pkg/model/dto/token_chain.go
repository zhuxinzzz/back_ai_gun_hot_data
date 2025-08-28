package dto

import (
	"back_ai_gun_data/utils"
	"time"

	"gorm.io/gorm"
)

// TokenChain 代币与区块链关系表
type TokenChain struct {
	UUID            string `gorm:"primaryKey;column:uuid;type:varchar(36)" json:"uuid"`
	TokenID         string `gorm:"column:token_id;type:varchar(36);not null" json:"token_id"`         // 代币UUID
	ChainID         string `gorm:"column:chain_id;type:varchar(36);not null" json:"chain_id"`         // 区块链UUID
	ContractAddress string `gorm:"column:contract_address;type:varchar(255)" json:"contract_address"` // 合约地址
	Decimals        int    `gorm:"column:decimals;type:int;default:18" json:"decimals"`               // 代币精度
	CreatedAt       int64  `gorm:"column:created_at" json:"created_at"`                               // 创建时间戳(毫秒)
	UpdatedAt       int64  `gorm:"column:updated_at" json:"updated_at"`                               // 更新时间戳(毫秒)
	DeletedAt       int64  `gorm:"column:deleted_at" json:"deleted_at"`                               // 删除时间戳(毫秒)
}

func (TokenChain) TableName() string {
	return "token_chains"
}

// BeforeCreate 创建前钩子
func (tc *TokenChain) BeforeCreate(tx *gorm.DB) error {
	// 自动生成UUID
	if tc.UUID == "" {
		tc.UUID = utils.GenerateUUIDV7()
	}

	if tc.CreatedAt == 0 {
		tc.CreatedAt = time.Now().UnixMilli()
	}
	if tc.UpdatedAt == 0 {
		tc.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (tc *TokenChain) BeforeUpdate(tx *gorm.DB) error {
	tc.UpdatedAt = time.Now().UnixMilli()
	return nil
}

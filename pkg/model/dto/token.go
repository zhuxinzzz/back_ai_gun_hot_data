package dto

import (
	"back_ai_gun_data/utils"
	"time"

	"gorm.io/gorm"
)

// Token 代币信息表
type Token struct {
	UUID            string  `gorm:"primaryKey;column:uuid;type:varchar(36)" json:"uuid"`
	TokenID         string  `gorm:"column:token_id;type:varchar(100)" json:"token_id"`                              // 各平台的代币ID，可为空
	Name            string  `gorm:"column:name;type:varchar(255);not null" json:"name"`                             // 代币全名，如 Bitcoin
	Symbol          string  `gorm:"column:symbol;type:varchar(50);not null" json:"symbol"`                          // 代币简称，如 BTC
	Volume24Hour    float64 `gorm:"column:volume_24hour;type:decimal(20,8);default:0" json:"volume_24hour"`         // 24小时交易量,整个市场
	MarketCap24Hour float64 `gorm:"column:market_cap_24hour;type:decimal(20,8);default:0" json:"market_cap_24hour"` // 市值，整个市场
	LogoURL         string  `gorm:"column:logo_url;type:varchar(500)" json:"logo_url"`                              // 图标URL，可以为空
	SourceAPI       string  `gorm:"column:source_api;type:varchar(100)" json:"source_api"`                          // 数据来源API，可以为空
	CreatedAt       int64   `gorm:"column:created_at" json:"created_at"`                                            // 创建时间戳(毫秒)
	UpdatedAt       int64   `gorm:"column:updated_at" json:"updated_at"`                                            // 更新时间戳(毫秒)
	DeletedAt       int64   `gorm:"column:deleted_at" json:"deleted_at"`                                            // 删除时间戳(毫秒)
}

func (Token) TableName() string {
	return "tokens"
}

// BeforeCreate 创建前钩子
func (t *Token) BeforeCreate(tx *gorm.DB) error {
	// 自动生成UUID
	if t.UUID == "" {
		t.UUID = utils.GenerateUUIDV7()
	}

	if t.CreatedAt == 0 {
		t.CreatedAt = time.Now().UnixMilli()
	}
	if t.UpdatedAt == 0 {
		t.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (t *Token) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now().UnixMilli()
	return nil
}

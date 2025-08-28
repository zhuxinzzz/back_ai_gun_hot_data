package dto

import (
	"time"

	"gorm.io/gorm"
)

// Chain 区块链网络表
type Chain struct {
	ID                     string    `gorm:"primaryKey;column:id;type:uuid" json:"id"`
	CreatedAt              time.Time `gorm:"column:created_at;type:timestamp(3)" json:"created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at;type:timestamp(3)" json:"updated_at"`
	IsDeleted              bool      `gorm:"column:is_deleted;type:boolean;default:false" json:"is_deleted"`
	IsActive               *bool     `gorm:"column:is_active;type:boolean;default:true" json:"is_active"`
	Slug                   string    `gorm:"column:slug;type:text;not null;unique" json:"slug"`
	NetworkID              *string   `gorm:"column:network_id;type:text" json:"network_id"`
	Type                   *string   `gorm:"column:type;type:text" json:"type"`
	MainToken              *string   `gorm:"column:main_token;type:jsonb" json:"main_token"`
	Name                   string    `gorm:"column:name;type:text;not null" json:"name"`
	Symbol                 *string   `gorm:"column:symbol;type:text" json:"symbol"`
	Rpcs                   *string   `gorm:"column:rpcs;type:jsonb" json:"rpcs"`
	Logo                   *string   `gorm:"column:logo;type:text" json:"logo"`
	Slip44                 *string   `gorm:"column:slip44;type:text" json:"slip44"`
	Explorers              *string   `gorm:"column:explorers;type:jsonb" json:"explorers"`
	OkxChainIndex          *string   `gorm:"column:okx_chain_index;type:text" json:"okx_chain_index"`
	Mapping                *string   `gorm:"column:mapping;type:varchar(65535)" json:"mapping"`
	CoinGeckoChainName     *string   `gorm:"column:coin_gecko_chain_name;type:text" json:"coin_gecko_chain_name"`
	CoinMarketCapChainName *string   `gorm:"column:coin_market_cap_chain_name;type:text" json:"coin_market_cap_chain_name"`
}

func (Chain) TableName() string {
	return "chain"
}

// BeforeCreate 创建前钩子
func (c *Chain) BeforeCreate(tx *gorm.DB) error {
	// 设置默认值
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (c *Chain) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

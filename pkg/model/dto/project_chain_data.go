package dto

import (
	"time"
)

// ProjectChainData 项目链数据模型
type ProjectChainData struct {
	ID                   string    `json:"id" gorm:"primaryKey;column:id;type:uuid"`
	CreatedAt            time.Time `json:"created_at" gorm:"column:created_at;type:timestamp(3)"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp(3)"`
	IsVisible            bool      `json:"is_visible" gorm:"column:is_visible;type:boolean;default:true"`
	IsDeleted            bool      `json:"is_deleted" gorm:"column:is_deleted;type:boolean;default:false"`
	EntityID             *string   `json:"entity_id" gorm:"column:entity_id;type:uuid"`
	ProjectID            *string   `json:"project_id" gorm:"column:project_id;type:uuid"`
	ChainID              *string   `json:"chain_id" gorm:"column:chain_id;type:uuid;not null"`
	ContractAddress      string    `json:"contract_address" gorm:"column:contract_address;type:text;not null"`
	Type                 *string   `json:"type" gorm:"column:type;type:text"`
	Standard             *string   `json:"standard" gorm:"column:standard;type:text"`
	Decimals             *int      `json:"decimals" gorm:"column:decimals;type:integer"`
	Version              *string   `json:"version" gorm:"column:version;type:text"`
	Name                 *string   `json:"name" gorm:"column:name;type:text"`
	Symbol               *string   `json:"symbol" gorm:"column:symbol;type:text"`
	Logo                 *string   `json:"logo" gorm:"column:logo;type:text"`
	LifiCoinKey          *string   `json:"lifi_coin_key" gorm:"column:lifi_coin_key;type:text"`
	TradingVolume24Hours *float64  `json:"trading_volume_24_hours" gorm:"column:volume_24h;type:float8"`
	MarketCap24Hours     *float64  `json:"market_cap_24_hours" gorm:"column:market_cap;type:float8"`
	Price24Hours         *float64  `json:"price_24_hours" gorm:"column:price_usd;type:float8"`
	Description          string    `gorm:"column:description;type:text" json:"description"`
}

package dto

import (
	"time"

	"gorm.io/gorm"
)

// Intelligence 情报表 - 根据SQL表结构定义
type Intelligence struct {
	ID          string    `gorm:"primaryKey;column:id;type:uuid" json:"id"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp(3);default:CURRENT_TIMESTAMP(3)" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp(3);default:CURRENT_TIMESTAMP(3)" json:"updated_at"`
	IsDeleted   bool      `gorm:"column:is_deleted;type:boolean;default:false" json:"is_deleted"`
	IsVisible   bool      `gorm:"column:is_visible;type:boolean;default:true" json:"is_visible"`
	IsValuable  bool      `gorm:"column:is_valuable;type:boolean;default:false" json:"is_valuable"`
	PublishedAt time.Time `gorm:"column:published_at;type:timestamp(3)" json:"published_at"`
	Type        string    `gorm:"column:type;type:text;not null" json:"type"`
	Subtype     *string   `gorm:"column:subtype;type:text" json:"subtype"`
	SourceID    string    `gorm:"column:source_id;type:uuid;not null" json:"source_id"`
	SourceURL   string    `gorm:"column:source_url;type:text;not null" json:"source_url"`
	Title       *string   `gorm:"column:title;type:text" json:"title"`
	Content     *string   `gorm:"column:content;type:text" json:"content"`
	Abstract    *string   `gorm:"column:abstract;type:text" json:"abstract"`
	ExtraDatas  *string   `gorm:"column:extra_datas;type:jsonb" json:"extra_datas"`
	Tags        *string   `gorm:"column:tags;type:jsonb" json:"tags"`
	Medias      *string   `gorm:"column:medias;type:jsonb" json:"medias"`
	Analyzed    *string   `gorm:"column:analyzed;type:jsonb" json:"analyzed"`
	Score       float64   `gorm:"column:score;type:double precision;default:0.0" json:"score"`
}

func (Intelligence) TableName() string {
	return "intelligence"
}

// BeforeCreate 创建前钩子
func (i *Intelligence) BeforeCreate(tx *gorm.DB) error {
	if i.CreatedAt.IsZero() {
		i.CreatedAt = time.Now()
	}
	if i.UpdatedAt.IsZero() {
		i.UpdatedAt = time.Now()
	}
	if i.PublishedAt.IsZero() {
		i.PublishedAt = time.Now()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (i *Intelligence) BeforeUpdate(tx *gorm.DB) error {
	i.UpdatedAt = time.Now()
	return nil
}

// EntityIntelligence 情报实体关联表
type EntityIntelligence struct {
	ID             string    `gorm:"primaryKey;column:id;type:uuid" json:"id"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp(3);default:CURRENT_TIMESTAMP(3)" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp(3);default:CURRENT_TIMESTAMP(3)" json:"updated_at"`
	IsDeleted      bool      `gorm:"column:is_deleted;type:boolean;default:false" json:"is_deleted"`
	IntelligenceID string    `gorm:"column:intelligence_id;type:uuid;not null" json:"intelligence_id"`
	EntityID       string    `gorm:"column:entity_id;type:uuid;not null" json:"entity_id"`
	Type           *string   `gorm:"column:type;type:text" json:"type"`
	ExtraData      *string   `gorm:"column:extra_data;type:jsonb" json:"extra_data"`
}

func (EntityIntelligence) TableName() string {
	return "entity_intelligence"
}

// BeforeCreate 创建前钩子
func (ei *EntityIntelligence) BeforeCreate(tx *gorm.DB) error {
	if ei.CreatedAt.IsZero() {
		ei.CreatedAt = time.Now()
	}
	if ei.UpdatedAt.IsZero() {
		ei.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (ei *EntityIntelligence) BeforeUpdate(tx *gorm.DB) error {
	ei.UpdatedAt = time.Now()
	return nil
}

package dto

import (
	"time"

	"gorm.io/gorm"
)

// Entity 实体表
type Entity struct {
	ID             string    `gorm:"primaryKey;column:id;type:uuid" json:"id"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp(3);default:CURRENT_TIMESTAMP(3)" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp(3);default:CURRENT_TIMESTAMP(3)" json:"updated_at"`
	IsDeleted      bool      `gorm:"column:is_deleted;type:boolean;default:false" json:"is_deleted"`
	IsVisible      bool      `gorm:"column:is_visible;type:boolean;default:true" json:"is_visible"`
	IsTest         bool      `gorm:"column:is_test;type:boolean;default:true" json:"is_test"`
	Name           string    `gorm:"column:name;type:text;not null" json:"name"`
	Slug           string    `gorm:"column:slug;type:text;not null;unique" json:"slug"`
	Type           string    `gorm:"column:type;type:text;not null" json:"type"`
	InfluenceLevel *string   `gorm:"column:influence_level;type:text" json:"influence_level"`
	InfluenceScore float64   `gorm:"column:influence_score;type:double precision;default:0" json:"influence_score"`
	Source         *string   `gorm:"column:source;type:text" json:"source"`
	Locations      *string   `gorm:"column:locations;type:jsonb" json:"locations"`
	Description    *string   `gorm:"column:description;type:text" json:"description"`
	Avatar         *string   `gorm:"column:avatar;type:text" json:"avatar"`
	ExtraData      *string   `gorm:"column:extra_data;type:jsonb" json:"extra_data"`
	Version        *string   `gorm:"column:version;type:text" json:"version"`
	Subtype        *string   `gorm:"column:subtype;type:text" json:"subtype"`
}

func (Entity) TableName() string {
	return "entity"
}

// BeforeCreate 创建前钩子
func (e *Entity) BeforeCreate(tx *gorm.DB) error {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (e *Entity) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}

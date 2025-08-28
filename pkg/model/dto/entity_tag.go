package dto

import (
	"time"

	"gorm.io/gorm"
)

// EntityTag 实体标签关联表
type EntityTag struct {
	ID        string    `gorm:"primaryKey;column:id;type:uuid" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp(3);default:CURRENT_TIMESTAMP(3)" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp(3);default:CURRENT_TIMESTAMP(3)" json:"updated_at"`
	IsDeleted bool      `gorm:"column:is_deleted;type:boolean;default:false" json:"is_deleted"`
	EntityID  string    `gorm:"column:entity_id;type:uuid;not null" json:"entity_id"`
	TagID     string    `gorm:"column:tag_id;type:uuid;not null" json:"tag_id"`
	Type      *string   `gorm:"column:type;type:text" json:"type"`
}

func (EntityTag) TableName() string {
	return "entity_tag"
}

// BeforeCreate 创建前钩子
func (et *EntityTag) BeforeCreate(tx *gorm.DB) error {
	if et.CreatedAt.IsZero() {
		et.CreatedAt = time.Now()
	}
	if et.UpdatedAt.IsZero() {
		et.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (et *EntityTag) BeforeUpdate(tx *gorm.DB) error {
	et.UpdatedAt = time.Now()
	return nil
}

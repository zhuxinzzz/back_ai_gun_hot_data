package dto

import (
	"time"

	"gorm.io/gorm"
)

// Tag 标签表
type Tag struct {
	ID        string    `gorm:"primaryKey;column:id;type:uuid" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp(3);default:CURRENT_TIMESTAMP(3)" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp(3);default:CURRENT_TIMESTAMP(3)" json:"updated_at"`
	IsDeleted bool      `gorm:"column:is_deleted;type:boolean;default:false" json:"is_deleted"`
	IsVisible bool      `gorm:"column:is_visible;type:boolean;default:true" json:"is_visible"`
	Slug      *string   `gorm:"column:slug;type:text;unique" json:"slug"`
	Name      *string   `gorm:"column:name;type:text" json:"name"`
}

func (Tag) TableName() string {
	return "tag"
}

// BeforeCreate 创建前钩子
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (t *Tag) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}

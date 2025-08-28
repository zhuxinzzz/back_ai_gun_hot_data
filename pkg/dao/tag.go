package dao

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"errors"

	"gorm.io/gorm"
)

// GetTagBySlug 根据slug获取tag
func GetTagBySlug(slug string) (*dto.Tag, error) {
	var tag dto.Tag
	result := GetDB().Where("slug = ? AND is_deleted = false", slug).First(&tag)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		lr.E().Errorf("Failed to get tag by slug %s: %v", slug, result.Error)
		return nil, result.Error
	}
	return &tag, nil
}

// CreateTag 创建新tag（如果不存在）
func CreateTag(tag *dto.Tag) error {
	// 先检查是否已存在
	existingTag, err := GetTagBySlug(*tag.Slug)
	if err != nil {
		return err
	}

	if existingTag != nil {
		lr.I().Infof("Tag already exists with slug: %s", *tag.Slug)
		return nil
	}

	result := GetDB().Create(tag)
	if result.Error != nil {
		lr.E().Errorf("Failed to create tag: %v", result.Error)
		return result.Error
	}

	lr.I().Infof("Created new tag with slug: %s", *tag.Slug)
	return nil
}

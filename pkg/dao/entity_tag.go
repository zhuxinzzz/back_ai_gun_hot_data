package dao

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"errors"

	"gorm.io/gorm"
)

// CreateEntityTag 创建entity_tag关联（如果不存在）
func CreateEntityTag(entityTag *dto.EntityTag) error {
	// 先检查是否已存在相同的关联
	var existingEntityTag dto.EntityTag
	result := GetDB().Where("entity_id = ? AND tag_id = ? AND is_deleted = false",
		entityTag.EntityID, entityTag.TagID).First(&existingEntityTag)

	if result.Error == nil {
		lr.I().Infof("EntityTag association already exists: EntityID=%s, TagID=%s",
			entityTag.EntityID, entityTag.TagID)
		return nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		lr.E().Errorf("Failed to check existing entity tag: %v", result.Error)
		return result.Error
	}

	// 创建新关联
	result = GetDB().Create(entityTag)
	if result.Error != nil {
		lr.E().Errorf("Failed to create entity tag: %v", result.Error)
		return result.Error
	}

	//lr.I().Infof("Created new entity tag: EntityID=%s, TagID=%s",
	//	entityTag.EntityID, entityTag.TagID)
	return nil
}

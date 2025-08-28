package dao

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"errors"

	"gorm.io/gorm"
)

// GetEntityBySlugAndType 根据slug和type获取entity
func GetEntityBySlugAndType(slug, entityType string) (*dto.Entity, error) {
	var entity dto.Entity
	result := GetDB().Where("slug = ? AND type = ? AND is_deleted = false", slug, entityType).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		lr.E().Errorf("Failed to get entity by slug %s and type %s: %v", slug, entityType, result.Error)
		return nil, result.Error
	}
	return &entity, nil
}

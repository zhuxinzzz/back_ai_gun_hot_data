package dao

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/utils"
	"errors"

	"gorm.io/gorm"
)

func CreateIntelligence(intelligence *dto.Intelligence) error {
	if intelligence.ID == "" {
		intelligence.ID = utils.GenerateUUIDV7()
	}

	result := GetDB().Create(intelligence)
	if result.Error != nil {
		lr.E().Errorf("Failed to create intelligence: %v", result.Error)
		return result.Error
	}

	lr.I().Infof("Created intelligence: %s", intelligence.ID)
	return nil
}

// GetIntelligenceByID 根据ID获取情报
func GetIntelligenceByID(id string) (*dto.Intelligence, error) {
	var intelligence dto.Intelligence
	result := GetDB().Where("id = ? AND is_deleted = false", id).First(&intelligence)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		lr.E().Errorf("Failed to get intelligence by id %s: %v", id, result.Error)
		return nil, result.Error
	}
	return &intelligence, nil
}

// CreateEntityIntelligence 创建情报实体关联
func CreateEntityIntelligence(entityIntelligence *dto.EntityIntelligence) error {
	if entityIntelligence.ID == "" {
		entityIntelligence.ID = utils.GenerateUUIDV7()
	}

	// 检查是否已存在相同的关联
	var existing dto.EntityIntelligence
	result := GetDB().Where("intelligence_id = ? AND entity_id = ? AND is_deleted = false",
		entityIntelligence.IntelligenceID, entityIntelligence.EntityID).First(&existing)

	if result.Error == nil {
		lr.I().Infof("EntityIntelligence association already exists: IntelligenceID=%s, EntityID=%s",
			entityIntelligence.IntelligenceID, entityIntelligence.EntityID)
		return nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		lr.E().Errorf("Failed to check existing entity intelligence: %v", result.Error)
		return result.Error
	}

	// 创建新关联
	result = GetDB().Create(entityIntelligence)
	if result.Error != nil {
		lr.E().Errorf("Failed to create entity intelligence: %v", result.Error)
		return result.Error
	}

	lr.I().Infof("Created entity intelligence: IntelligenceID=%s, EntityID=%s",
		entityIntelligence.IntelligenceID, entityIntelligence.EntityID)
	return nil
}

// GetEntitiesByIntelligenceID 根据情报ID获取关联的实体列表
func GetEntitiesByIntelligenceID(intelligenceID string) ([]dto.Entity, error) {
	var entities []dto.Entity

	result := GetDB().Joins("JOIN entity_intelligence ei ON entity.id = ei.entity_id").
		Where("ei.intelligence_id = ? AND ei.is_deleted = false AND entity.is_deleted = false", intelligenceID).
		Find(&entities)

	if result.Error != nil {
		lr.E().Errorf("Failed to get entities by intelligence id %s: %v", intelligenceID, result.Error)
		return nil, result.Error
	}

	return entities, nil
}

// GetIntelligencesByEntityID 根据实体ID获取关联的情报列表
func GetIntelligencesByEntityID(entityID string) ([]dto.Intelligence, error) {
	var intelligences []dto.Intelligence

	result := GetDB().Joins("JOIN entity_intelligence ei ON intelligence.id = ei.intelligence_id").
		Where("ei.entity_id = ? AND ei.is_deleted = false AND intelligence.is_deleted = false", entityID).
		Find(&intelligences)

	if result.Error != nil {
		lr.E().Errorf("Failed to get intelligences by entity id %s: %v", entityID, result.Error)
		return nil, result.Error
	}

	return intelligences, nil
}

// GetAllActiveIntelligences 获取所有活跃的情报
func GetAllActiveIntelligences() ([]*dto.Intelligence, error) {
	var intelligences []*dto.Intelligence

	result := GetDB().Where("is_deleted = false AND is_visible = true").
		Order("is_valuable DESC, published_at DESC").
		Find(&intelligences)

	if result.Error != nil {
		lr.E().Errorf("Failed to get all active intelligences: %v", result.Error)
		return nil, result.Error
	}

	return intelligences, nil
}

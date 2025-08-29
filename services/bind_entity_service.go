package services

import (
	"back_ai_gun_data/pkg/dao"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"back_ai_gun_data/utils"
)

const (
	// EntityTypeProject 项目类型实体
	EntityTypeProject = "project"
	// EntityTagTypeAlias 实体标签类型：别名
	EntityTagTypeAlias = "alias"
)

func ProcessTagAssociation(entityID, projectName string) error {
	// 查找tag
	tag, err := dao.GetTagBySlug(projectName)
	if err != nil {
		lr.E().Errorf("Failed to get tag by slug: %v", err)
		return err
	}

	var tagID string
	if tag == nil {
		// 创建新tag
		newTag := &dto.Tag{
			ID:        utils.GenerateUUIDV7(),
			Slug:      &projectName,
			Name:      &projectName,
			IsVisible: true,
			IsDeleted: false,
		}

		if err := dao.CreateTag(newTag); err != nil {
			lr.E().Errorf("Failed to create tag: %v", err)
			return err
		}
		tagID = newTag.ID
		lr.I().Infof("创建新tag: %s", projectName)
	} else {
		tagID = tag.ID
	}

	// 创建entity_tag关联
	entityTag := &dto.EntityTag{
		ID:        utils.GenerateUUIDV7(),
		EntityID:  entityID,
		TagID:     tagID,
		Type:      &[]string{EntityTagTypeAlias}[0], // type是alias
		IsDeleted: false,
	}

	if err := dao.CreateEntityTag(entityTag); err != nil {
		lr.E().Errorf("Failed to create entity tag: %v", err)
		return err
	}

	lr.I().Infof("创建entity_tag关联，Entity ID: %s, Tag ID: %s", entityID, tagID)
	return nil
}

// BindAllEntities 绑定所有实体 - 扁平化实现
func BindAllEntities() error {
	// 第一步：清空project_chain_data的entity_id
	//if err := dao.ClearProjectChainDataEntityID(); err != nil {
	//	lr.E().Errorf("清空project_chain_data的entity_id失败: %v", err)
	//	return err
	//}
	//lr.I().Info("已清空project_chain_data的entity_id")

	// 第二步：分批处理绑定
	const batchSize = 800
	offset := 0
	totalProcessed := 0
	totalBound := 0

	for {
		// 获取一批没有entity_id的project_chain_data
		projectChains, err := dao.GetProjectChainDataWithoutEntityID(offset, batchSize)
		if err != nil {
			lr.E().Errorf("获取project_chain_data失败: %v", err)
			return err
		}

		if len(projectChains) == 0 {
			break // 没有更多数据
		}

		//lr.I().Infof("处理第 %d-%d 条记录", offset+1, offset+len(projectChains))
		// 处理这一批数据
		boundCount := 0
		for _, projectChain := range projectChains {
			if projectChain.Name == nil {
				lr.I().Infof("跳过没有name的记录: %s", projectChain.ID)
				continue
			}
			// 标准化名称（小写去空格）
			normalizedName := utils.NormalizeName(*projectChain.Name)

			// 查找entity（type是project）
			entity, err := dao.GetEntityBySlugAndType(normalizedName, EntityTypeProject)
			if err != nil {
				lr.E().Errorf("查找entity失败，ID: %s, Name: %s, Error: %v", projectChain.ID, *projectChain.Name, err)
				continue
			}
			if entity == nil {
				lr.I().Infof("未找到对应的entity，Name: %s, NormalizedName: %s", *projectChain.Name, normalizedName)
				continue
			}

			// 更新project_chain_data的entity_id
			if err := dao.UpdateProjectChainDataEntityID(projectChain.ID, entity.ID); err != nil {
				lr.E().Errorf("更新entity_id失败，ID: %s, Error: %v", projectChain.ID, err)
				continue
			}

			// 处理tag关联
			if err := ProcessTagAssociation(entity.ID, *projectChain.Name); err != nil {
				lr.I().Infof("处理tag关联失败，但不影响主流程: %v", err)
			}

			lr.I().Infof("成功绑定实体，ProjectChainData ID: %s, Entity ID: %s", projectChain.ID, entity.ID)
			boundCount++
		}

		totalProcessed += len(projectChains)
		totalBound += boundCount
		offset += batchSize

		lr.I().Infof("批次处理完成，绑定 %d/%d 条记录", boundCount, len(projectChains))
	}

	lr.I().Infof("所有实体绑定完成，总共处理 %d 条记录，成功绑定 %d 条", totalProcessed, totalBound)
	return nil
}

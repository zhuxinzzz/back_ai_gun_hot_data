package dao

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/dto"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func UpdateProjectChainData(data *dto.ProjectChainData) error {
	result := GetDB().Save(data)
	if result.Error != nil {
		lr.E().Errorf("Failed to update project chain data: %v", result.Error)
		return result.Error
	}

	//lr.I().Infof("Successfully updated project chain data with ID: %s", data.ID)
	return nil
}

func UpdateProjectChainDataLogo(id string, logoURL string) error {
	result := GetDB().Model(&dto.ProjectChainData{}).Where("id = ?", id).Updates(map[string]interface{}{
		"logo":       logoURL,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		lr.E().Errorf("Failed to update project chain data logo: %v", result.Error)
		return result.Error
	}

	return nil
}

func GetProjectChainDataByChainIDAndContractAddress(chainID, contractAddress string) (*dto.ProjectChainData, error) {
	var data dto.ProjectChainData
	result := GetDB().Where("chain_id = ? AND contract_address = ? AND is_deleted = false", chainID, contractAddress).First(&data)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			lr.I().Infof("No project chain data found for chain_id: %s, contract_address: %s", chainID, contractAddress)
			return nil, nil
		}
		lr.E().Errorf("Failed to get project chain data: %v", result.Error)
		return nil, result.Error
	}

	return &data, nil
}

func GetProjectChainDataByChainIDAndAddresses(chainID string, addresses []string) (map[string]*dto.ProjectChainData, error) {
	if len(addresses) == 0 {
		return make(map[string]*dto.ProjectChainData), nil
	}

	var dataList []dto.ProjectChainData
	result := GetDB().Where("chain_id = ? AND contract_address IN ? AND is_deleted = false", chainID, addresses).Find(&dataList)
	if result.Error != nil {
		lr.E().Errorf("Failed to get project chain data by chain_id: %s, addresses: %v, error: %v", chainID, addresses, result.Error)
		return nil, result.Error
	}

	// 构建地址到数据的映射
	dataMap := make(map[string]*dto.ProjectChainData)
	for i := range dataList {
		dataMap[dataList[i].ContractAddress] = &dataList[i]
	}

	//lr.I().Infof("Successfully found %d project chain data records for chain_id: %s", len(dataMap), chainID)
	return dataMap, nil
}

// GetProjectChainDataByNamesAndAddresses 根据币名（模糊匹配）和地址查询项目链数据
func GetProjectChainDataByNamesAndAddresses(names []string, addresses []string) ([]*dto.ProjectChainData, error) {
	if len(names) == 0 && len(addresses) == 0 {
		return []*dto.ProjectChainData{}, nil
	}

	var dataList []dto.ProjectChainData
	dataList = make([]dto.ProjectChainData, 0, len(addresses))
	query := GetDB().Where("is_deleted = false")

	// 添加币名模糊匹配条件
	if len(names) > 0 {
		var nameConditions []string
		var nameArgs []interface{}
		for _, name := range names {
			nameConditions = append(nameConditions, "name ILIKE ?")
			nameArgs = append(nameArgs, "%"+name+"%")
		}
		query = query.Where("("+strings.Join(nameConditions, " OR ")+")", nameArgs...)
	}

	// 添加地址精确匹配条件
	if len(addresses) > 0 {
		query = query.Where("contract_address IN ?", addresses)
	}

	result := query.Find(&dataList)
	if result.Error != nil {
		lr.E().Errorf("Failed to get project chain data by names: %v, addresses: %v, error: %v", names, addresses, result.Error)
		return nil, result.Error
	}

	// 转换为指针切片
	resultList := make([]*dto.ProjectChainData, len(dataList))
	for i := range dataList {
		resultList[i] = &dataList[i]
	}

	return resultList, nil
}

// GetUnfollowedProjectChainData 获取未关注的项目链数据（is_follow为false或不存在）
func GetUnfollowedProjectChainData(names []string, addresses []string) ([]*dto.ProjectChainData, error) {
	if len(names) == 0 && len(addresses) == 0 {
		return []*dto.ProjectChainData{}, nil
	}

	var dataList []dto.ProjectChainData
	query := GetDB().Where("is_deleted = false AND (is_follow = false OR is_follow IS NULL)")

	// 添加币名模糊匹配条件
	if len(names) > 0 {
		var nameConditions []string
		var nameArgs []interface{}
		for _, name := range names {
			nameConditions = append(nameConditions, "name ILIKE ?")
			nameArgs = append(nameArgs, "%"+name+"%")
		}
		query = query.Where("("+strings.Join(nameConditions, " OR ")+")", nameArgs...)
	}

	// 添加地址精确匹配条件
	if len(addresses) > 0 {
		query = query.Where("contract_address IN ?", addresses)
	}

	result := query.Find(&dataList)
	if result.Error != nil {
		lr.E().Errorf("Failed to get unfollowed project chain data by names: %v, addresses: %v, error: %v", names, addresses, result.Error)
		return nil, result.Error
	}

	// 转换为指针切片
	resultList := make([]*dto.ProjectChainData, len(dataList))
	for i := range dataList {
		resultList[i] = &dataList[i]
	}

	return resultList, nil
}

func CreateProjectChainData(data *dto.ProjectChainData) error {
	result := pgDB.Create(data)
	if result.Error != nil {
		lr.E().Errorf("Failed to create project chain data: %v", result.Error)
		return result.Error
	}

	return nil
}

func BatchCreateProjectChainData(dataList []*dto.ProjectChainData) error {
	result := GetDB().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "chain_id"},
			{Name: "contract_address"}},
		DoNothing: true,
	}).CreateInBatches(dataList, 100)

	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func BatchUpdateProjectChainDataLogo(updates []LogoUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// 使用事务确保批量更新的原子性
	tx := GetDB().Begin()
	if tx.Error != nil {
		lr.E().Errorf("Failed to begin transaction for batch logo update: %v", tx.Error)
		return tx.Error
	}

	var errors []string
	successCount := 0

	for _, update := range updates {
		result := tx.Model(&dto.ProjectChainData{}).Where("id = ?", update.ID).Updates(map[string]interface{}{
			"logo":       update.LogoURL,
			"updated_at": time.Now(),
		})
		if result.Error != nil {
			errors = append(errors, fmt.Sprintf("ID %s: %v", update.ID, result.Error))
			lr.E().Errorf("Failed to update logo for ID %s: %v", update.ID, result.Error)
			continue
		}
		successCount++
	}

	if len(errors) > 0 {
		tx.Rollback()
		lr.E().Errorf("Batch logo update failed with %d errors: %v", len(errors), errors)
		return fmt.Errorf("batch logo update failed: %v", errors)
	}

	if err := tx.Commit().Error; err != nil {
		lr.E().Errorf("Failed to commit batch logo update transaction: %v", err)
		return err
	}

	lr.I().Infof("Successfully batch updated %d logo records", successCount)
	return nil
}

func GetProjectChainDataByID(id string) (*dto.ProjectChainData, error) {
	var data dto.ProjectChainData
	result := GetDB().Where("id = ? AND is_deleted = false", id).First(&data)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			lr.I().Infof("No project chain data found for ID: %s", id)
			return nil, nil
		}
		lr.E().Errorf("Failed to get project chain data by ID: %v", result.Error)
		return nil, result.Error
	}

	return &data, nil
}

func GetProjectChainDataByProjectID(projectID string) ([]dto.ProjectChainData, error) {
	var dataList []dto.ProjectChainData
	result := GetDB().Where("project_id = ? AND is_deleted = false", projectID).Find(&dataList)
	if result.Error != nil {
		lr.E().Errorf("Failed to get project chain data by project ID: %v", result.Error)
		return nil, result.Error
	}

	return dataList, nil
}

func DeleteProjectChainData(id string) error {
	result := GetDB().Model(&dto.ProjectChainData{}).Where("id = ?", id).Update("is_deleted", true)
	if result.Error != nil {
		lr.E().Errorf("Failed to delete project chain data: %v", result.Error)
		return result.Error
	}

	//lr.I().Infof("Successfully deleted project chain data with ID: %s", id)
	return nil
}

func UpdateProjectChainDataMarketInfo(id string, price24Hours, tradingVolume24Hours, marketCap24Hours *float64) error {
	updates := map[string]interface{}{
		"price_usd":  price24Hours,
		"volume_24h": tradingVolume24Hours,
		"market_cap": marketCap24Hours,
		"updated_at": time.Now(),
	}

	result := GetDB().Model(&dto.ProjectChainData{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		lr.E().Errorf("Failed to update project chain data market info: %v", result.Error)
		return result.Error
	}

	return nil
}

func BatchUpdateProjectChainDataMarketInfo(updates []MarketInfoUpdate) error {
	var errors []string
	successCount := 0

	// 为每个更新使用单独的事务，避免一个失败影响整个批次
	for _, update := range updates {
		tx := GetDB().Begin()

		updates := map[string]interface{}{
			"price_usd":  update.Price24Hours,
			"volume_24h": update.TradingVolume24Hours,
			"market_cap": update.MarketCap24Hours,
			"updated_at": time.Now(),
		}

		result := tx.Model(&dto.ProjectChainData{}).Where("id = ?", update.ID).Updates(updates)
		if result.Error != nil {
			tx.Rollback()
			errors = append(errors, fmt.Sprintf("ID %s: %v", update.ID, result.Error))
			lr.E().Errorf("Failed to update project chain data ID %s: %v", update.ID, result.Error)
			continue
		}

		if err := tx.Commit().Error; err != nil {
			errors = append(errors, fmt.Sprintf("ID %s commit failed: %v", update.ID, err))
			lr.E().Errorf("Failed to commit update for ID %s: %v", update.ID, err)
			continue
		}

		successCount++
	}

	// 记录结果
	if len(errors) > 0 {
		lr.E().Warnf("Batch update completed with %d successes and %d errors: %v", successCount, len(errors), errors)
	} else {
		lr.I().Infof("Successfully batch updated %d project chain data market info", successCount)
	}

	// 如果有错误，返回错误信息但不中断程序
	if len(errors) > 0 {
		return fmt.Errorf("batch update completed with %d errors: %v", len(errors), errors)
	}

	return nil
}

type MarketInfoUpdate struct {
	ID                   string   `json:"id"`
	Price24Hours         *float64 `json:"price_24_hours"`
	TradingVolume24Hours *float64 `json:"trading_volume_24_hours"`
	MarketCap24Hours     *float64 `json:"market_cap_24_hours"`
}

type LogoUpdate struct {
	ID      string `json:"id"`
	LogoURL string `json:"logo_url"`
}

func GetProjectChains(offset, limit int) ([]*dto.ProjectChainData, error) {
	var projectChains []*dto.ProjectChainData
	result := GetDB().Offset(offset).Limit(limit).Find(&projectChains)
	if result.Error != nil {
		lr.E().Error(result.Error)
		return nil, result.Error
	}

	return projectChains, nil
}

func UpdateProjectChainDescription(projectChains []*dto.ProjectChainData) error {
	err := GetDB().Transaction(func(tx *gorm.DB) error {
		for _, projectChain := range projectChains {
			result := tx.Model(&dto.ProjectChainData{}).Where("id=?", projectChain.ID).
				Updates(map[string]any{
					"description": projectChain.Description,
					"updated_at":  time.Now(),
				})
			if result.Error != nil {
				lr.E().Error(result.Error)
				return result.Error
			}
		}
		return nil
	})

	return err
}

// UpdateProjectChainDataFields 更新指定字段
func UpdateProjectChainDataFields(id string, updates map[string]interface{}) error {
	// 自动添加updated_at
	updates["updated_at"] = time.Now()

	result := GetDB().Model(&dto.ProjectChainData{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		lr.E().Errorf("Failed to update project chain data fields: %v", result.Error)
		return result.Error
	}

	return nil
}

// BatchUpdateProjectChainDataFields 批量更新指定字段
func BatchUpdateProjectChainDataFields(updates []BatchUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// 使用事务确保批量更新的原子性
	tx := GetDB().Begin()
	if tx.Error != nil {
		lr.E().Errorf("Failed to begin transaction for batch update: %v", tx.Error)
		return tx.Error
	}

	var errors []string
	successCount := 0

	for _, update := range updates {
		// 自动添加updated_at
		update.Fields["updated_at"] = time.Now()

		result := tx.Model(&dto.ProjectChainData{}).Where("id = ?", update.ID).Updates(update.Fields)
		if result.Error != nil {
			errors = append(errors, fmt.Sprintf("ID %s: %v", update.ID, result.Error))
			lr.E().Errorf("Failed to update project chain data ID %s: %v", update.ID, result.Error)
			continue
		}
		successCount++
	}

	if len(errors) > 0 {
		tx.Rollback()
		lr.E().Errorf("Batch update failed with %d errors: %v", len(errors), errors)
		return fmt.Errorf("batch update failed: %v", errors)
	}

	if err := tx.Commit().Error; err != nil {
		lr.E().Errorf("Failed to commit batch update transaction: %v", err)
		return err
	}

	lr.I().Infof("Successfully batch updated %d project chain data records", successCount)
	return nil
}

// BatchUpdate 批量更新结构
type BatchUpdate struct {
	ID     string         `json:"id"`
	Fields map[string]any `json:"fields"`
}

// ClearProjectChainDataEntityID 清空project_chain_data的entity_id
func ClearProjectChainDataEntityID() error {
	result := GetDB().Model(&dto.ProjectChainData{}).Update("entity_id", nil)
	if result.Error != nil {
		lr.E().Errorf("Failed to clear project chain data entity_id: %v", result.Error)
		return result.Error
	}
	return nil
}

// UpdateProjectChainDataEntityID 更新project_chain_data的entity_id
func UpdateProjectChainDataEntityID(id, entityID string) error {
	result := GetDB().Model(&dto.ProjectChainData{}).Where("id = ?", id).Update("entity_id", entityID)
	if result.Error != nil {
		lr.E().Errorf("Failed to update project chain data entity_id: %v", result.Error)
		return result.Error
	}
	return nil
}

// GetProjectChainDataWithoutEntityID 获取没有entity_id的project_chain_data
func GetProjectChainDataWithoutEntityID(offset, limit int) ([]*dto.ProjectChainData, error) {
	var dataList []*dto.ProjectChainData
	result := GetDB().Where("entity_id IS NULL AND is_deleted = false").Offset(offset).Limit(limit).Find(&dataList)
	if result.Error != nil {
		lr.E().Errorf("Failed to get project chain data without entity_id: %v", result.Error)
		return nil, result.Error
	}
	return dataList, nil
}

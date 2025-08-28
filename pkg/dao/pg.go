package dao

import (
	"back_ai_gun_data/pkg/consts"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	pgDB *gorm.DB
)

// GetDB 统一的数据库获取函数 - 保持向后兼容
func GetDB() *gorm.DB {
	return pgDB
}

func GetPGDB() *gorm.DB {
	return pgDB
}

func connectPostgresSQL() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		consts.PG_HOST, consts.PG_PORT, consts.PG_USER, consts.PG_PASSWORD, consts.PG_NAME, consts.PG_SSLMODE)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		// 避免logger未初始化的问题
		fmt.Printf("PostgreSQL connection error: %v\n", err)
		return nil, fmt.Errorf("failed to connect to PostgreSQL database: %v", err)
	}

	return db, nil
}

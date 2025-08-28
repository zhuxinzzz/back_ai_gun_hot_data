package dao

import (
	"get_coin_info_v2/pkg/lr"
	"testing"
)

func TestPostgreSQLConnection(t *testing.T) {
	lr.Init()

	db := GetPGDB()
	if db == nil {
		t.Fatal("PostgreSQL database connection failed")
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	// 测试ping
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Failed to ping PostgreSQL database: %v", err)
	}

	t.Log("PostgreSQL connection test passed")
}

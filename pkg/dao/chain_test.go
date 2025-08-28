package dao

import (
	"encoding/csv"
	"get_coin_info_v2/pkg/lr"
	"get_coin_info_v2/pkg/model/dto"
	"io"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
)

// create table chain
// (
//
//	id                    uuid                                   not null
//	name                  text                                   not null,
//	coin_gecko_chain_name text
//
// 获取所有链信息，保存到文本文件id,name,coin_gecko_chain_name 是一个cvs,备后用
func TestSaveCoinGeckoToCSV(t *testing.T) {
	lr.Init()
	Init()
	//t.Skip("需要设置测试数据库连接")

	// 获取所有链信息
	chains, err := GetAllChains(GetPGDB())
	if err != nil {
		t.Errorf("GetAllChains failed: %v", err)
		return
	}

	if len(chains) == 0 {
		t.Log("No chains found in database")
		return
	}

	// 创建CSV文件
	filename := "chain_data.csv"
	file, err := os.Create(filename)
	if err != nil {
		t.Errorf("Failed to create CSV file: %v", err)
		return
	}
	defer file.Close()

	// 写入CSV头部
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := []string{"id", "name", "coin_gecko_chain_name"}
	if err := writer.Write(header); err != nil {
		t.Errorf("Failed to write CSV header: %v", err)
		return
	}

	// 写入数据行
	for _, chain := range chains {
		coinGeckoChainName := ""
		if chain.CoinGeckoChainName != nil {
			coinGeckoChainName = *chain.CoinGeckoChainName
		}

		row := []string{
			chain.ID,
			chain.Name,
			coinGeckoChainName,
		}

		if err := writer.Write(row); err != nil {
			t.Errorf("Failed to write CSV row: %v", err)
			return
		}
	}

	t.Logf("Successfully saved %d chains to %s", len(chains), filename)
}

// create table chain
//
//	id                    uuid                                   not null
//	name                  text                                   not null,
//	coin_gecko_chain_name text
//
// 从当前目录读取chain_data.csv,获得链信息的以上三个字段，遍历表。
// 如果id、name匹配，coin_gecko_chain_name 为空，就新增 coin_gecko_chain_name
func TestUpdateChainFromCSV(t *testing.T) {
	lr.Init()
	Init()
	//t.Skip("需要设置测试数据库连接")

	// 读取CSV文件
	filename := "chain_data.csv"
	file, err := os.Open(filename)
	if err != nil {
		t.Errorf("Failed to open CSV file: %v", err)
		return
	}
	defer file.Close()

	// 创建CSV读取器
	reader := csv.NewReader(file)

	// 读取表头
	header, err := reader.Read()
	if err != nil {
		t.Errorf("Failed to read CSV header: %v", err)
		return
	}

	// 验证表头格式
	expectedHeader := []string{"id", "name", "coin_gecko_chain_name"}
	if len(header) != len(expectedHeader) {
		t.Errorf("Invalid CSV header format, expected %d columns, got %d", len(expectedHeader), len(header))
		return
	}

	// 统计更新数量
	updatedCount := 0
	skippedCount := 0
	errorCount := 0

	// 逐行读取数据
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Errorf("Failed to read CSV record: %v", err)
			errorCount++
			continue
		}

		if len(record) != 3 {
			t.Logf("Skipping invalid record: %v", record)
			skippedCount++
			continue
		}

		id := record[0]
		name := record[1]
		coinGeckoChainName := record[2]

		// 跳过coin_gecko_chain_name为空的记录
		if coinGeckoChainName == "" {
			skippedCount++
			continue
		}

		// 根据ID和name查找匹配的记录
		var chain dto.Chain
		result := GetPGDB().Where("id = ? AND name = ? AND (coin_gecko_chain_name IS NULL OR coin_gecko_chain_name = '')", id, name).First(&chain)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				t.Logf("No matching record found for ID: %s, Name: %s", id, name)
				skippedCount++
				continue
			}
			t.Errorf("Database error for ID %s: %v", id, result.Error)
			errorCount++
			continue
		}

		// 安全更新：只更新coin_gecko_chain_name字段，避免覆盖其他字段
		updateResult := GetPGDB().Model(&dto.Chain{}).
			Where("id = ?", chain.ID).
			Updates(map[string]interface{}{
				"coin_gecko_chain_name": coinGeckoChainName,
				"updated_at":            time.Now(),
			})

		if updateResult.Error != nil {
			t.Errorf("Failed to update chain ID %s: %v", id, updateResult.Error)
			errorCount++
			continue
		}

		if updateResult.RowsAffected == 0 {
			t.Logf("No rows affected for chain ID: %s", id)
			skippedCount++
			continue
		}

		updatedCount++
		t.Logf("Updated chain ID: %s, Name: %s, CoinGeckoChainName: %s", id, name, coinGeckoChainName)
	}

	t.Logf("CSV processing completed - Updated: %d, Skipped: %d, Errors: %d", updatedCount, skippedCount, errorCount)
}

func TestGetChainByCoinGeckoChainName(t *testing.T) {
	lr.Init()
	Init()
	//t.Skip("需要设置测试数据库连接")

	chainName := "bnb"
	chain, err := GetChainByCoinGeckoChainName(chainName)
	if err != nil {
		t.Errorf("GetChainByCoinGeckoChainName failed: %v", err)
	}

	if chain != nil {
		t.Logf("Found chain: ID=%s, Name=%s", chain.ID, chain.Name)
	} else {
		t.Log("No chain found")
	}
}

func TestGetChainUUIDByCoinGeckoChainName(t *testing.T) {
	lr.Init()
	Init()
	// 这里需要设置测试数据库连接
	//t.Skip("需要设置测试数据库连接")

	chainName := "ethereum"
	uuid, err := GetChainUUIDByCoinGeckoChainName(chainName)
	if err != nil {
		t.Errorf("GetChainUUIDByCoinGeckoChainName failed: %v", err)
	}

	if uuid != "" {
		t.Logf("Found chain UUID: %s", uuid)
	} else {
		t.Log("No chain UUID found")
	}
}

func TestGetChainByUUID(t *testing.T) {
	lr.Init()
	Init()
	// 这里需要设置测试数据库连接
	//t.Skip("需要设置测试数据库连接")

	uuid := "fantom"
	chain, err := GetChainByUUID(GetPGDB(), uuid)
	if err != nil {
		t.Errorf("GetChainByUUID failed: %v", err)
	}

	if chain != nil {
		symbol := "nil"
		if chain.Symbol != nil {
			symbol = *chain.Symbol
		}
		t.Logf("Found chain: Name=%s, Symbol=%s", chain.Name, symbol)
	} else {
		t.Log("No chain found")
	}
}

package dao

import (
	"get_coin_info_v2/pkg/lr"
	"get_coin_info_v2/pkg/model/dto"
	"get_coin_info_v2/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateProjectChainData(t *testing.T) {
	// 这里需要设置测试数据库连接
	// 实际测试时需要 mock 数据库或使用测试数据库
	t.Skip("需要设置测试数据库连接")

	data := &dto.ProjectChainData{
		ID:              "test-id",
		ContractAddress: "0x1234567890abcdef",
		ChainID:         &[]string{"chain-1"}[0],
		Name:            &[]string{"Test Token"}[0],
		Symbol:          &[]string{"TEST"}[0],
		IsVisible:       true,
		IsDeleted:       false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := UpdateProjectChainData(data)
	if err != nil {
		t.Errorf("UpdateProjectChainData failed: %v", err)
	}
}

func TestGetProjectChainDataByChainIDAndContractAddress(t *testing.T) {
	// 这里需要设置测试数据库连接
	t.Skip("需要设置测试数据库连接")

	chainID := "chain-1"
	contractAddress := "0x1234567890abcdef"

	data, err := GetProjectChainDataByChainIDAndContractAddress(chainID, contractAddress)
	if err != nil {
		t.Errorf("GetProjectChainDataByChainIDAndContractAddress failed: %v", err)
	}

	if data != nil {
		t.Logf("Found data: ID=%s, Name=%s", data.ID, *data.Name)
	} else {
		t.Log("No data found")
	}
}

const version = "GET IN COIN GECKO"

func TestCreateProjectChainData(t *testing.T) {
	lr.Init()
	Init()

	chainId := "0197ac5a-b822-72a6-9519-ca49732a6b40"
	contractAddress := "0x039d2e8f097331278bd6c1415d839310e0d5ece4"

	decimals, _ := GetCoinGeckoCoinDecimalByID("aave-tusd-v1")
	coin, _ := GetCoinGeckoCoinByID("aave-tusd-v1")
	v := version

	data := &dto.ProjectChainData{
		ID:              utils.GenerateUUIDV7(),
		ContractAddress: contractAddress,
		ChainID:         &chainId,
		Name:            &coin.Name,
		Symbol:          &coin.Symbol,
		Version:         &v,
		Decimals:        &decimals,
		Logo:            &coin.Image.Small,
		IsVisible:       true,
		IsDeleted:       false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := CreateProjectChainData(data)
	if err != nil {
		t.Errorf("CreateProjectChainData failed: %v", err)
	}
}

func TestGetProjectChainDataWithoutEntityID(t *testing.T) {
	lr.Init()
	Init()
	offset := 1
	limit := 10

	res, err := GetProjectChainDataWithoutEntityID(offset, limit)
	assert.NoError(t, err)
	t.Log(utils.ToJson(res))
}

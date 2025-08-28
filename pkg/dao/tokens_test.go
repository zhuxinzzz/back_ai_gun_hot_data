package dao

import (
	"get_coin_info_v2/pkg/lr"
	"github.com/stretchr/testify/assert"
	"testing"
)

// func GetTokensBySymbol(mysqlDB *gorm.DB, symbol string) ([]dto.Token, error) {
func TestGetTokensBySymbol(t *testing.T) {
	lr.Init()
	Init()
	symbol := "btc"
	tokens, err := GetTokensBySymbol(symbol)
	assert.NoError(t, err)
	t.Logf("%+v", tokens)
}

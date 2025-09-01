package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/dao"
	"back_ai_gun_data/pkg/lr"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncShowedTokensToIntelligence(t *testing.T) {
	lr.Init()
	dao.Init()
	cache.Init()
	intelligenceID := "0198f0a9-0e77-721b-99df-b94e851375d1"

	err := SyncShowedTokensToIntelligence(intelligenceID)
	assert.NoError(t, err)
}

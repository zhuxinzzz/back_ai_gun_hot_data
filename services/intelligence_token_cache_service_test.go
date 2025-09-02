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
	intelligenceID := "019902fa-3cc7-71af-ad42-2c57caa4c25c"

	err := SyncShowedTokensToIntelligence(intelligenceID)
	assert.NoError(t, err)
}

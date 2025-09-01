package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model"
	"back_ai_gun_data/services/remote_service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_processRankingAndHotData(t *testing.T) {
	lr.Init()
	cache.Init()
	remote_service.Init()
	data := &model.MessageData{ID: "0198f0a9-0e77-721b-99df-b94e851375d1"}
	entities := analyzeEntities(data)
	ctx := context.Background()

	err := processRankingAndHotData(ctx, data, entities)
	assert.NoError(t, err)
}

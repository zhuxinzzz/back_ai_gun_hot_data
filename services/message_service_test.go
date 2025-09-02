package services

import (
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/dao"
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
	dao.Init()
	remote_service.Init()
	data := &model.IntelligenceMessage{
		BaseMessage: model.BaseMessage{
			ID: "019902fa-3cc7-71af-ad42-2c57caa4c25c",
		},
		Data: model.IntelligenceData{
			Entities: []model.IntelligenceEntity{},
		},
	}
	entities := analyzeEntities(data)
	ctx := context.Background()

	err := processRankingAndHotData(ctx, data, entities)
	assert.NoError(t, err)
}

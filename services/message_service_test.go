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
	dao.Init()
	cache.Init()
	remote_service.Init()
	data := &model.IntelligenceMessage{
		BaseMessage: model.BaseMessage{
			ID: "019902e7-3045-7d7c-88eb-e9330c62deac",
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

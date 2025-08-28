package services

import (
	"back_ai_gun_data/pkg/dao"
	"back_ai_gun_data/pkg/lr"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBindAllEntities(t *testing.T) {
	lr.Init()
	dao.Init()

	err := BindAllEntities()
	assert.NoError(t, err)
}

package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryTokensByName(t *testing.T) {
	lr.Init()
	name := "HOKK"
	chain := "Ethereum"

	res, err := QueryTokensByName(name, chain)
	assert.NoError(t, err)
	t.Log(utils.ToJson(res))
}

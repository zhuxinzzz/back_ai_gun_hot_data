package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryTokensByName(t *testing.T) {
	lr.Init()
	Init()
	name := "Kraken Wrapped BTC"
	chain := ""

	res, err := QueryTokensByName(name, chain)
	assert.NoError(t, err)
	t.Log(utils.ToJson(res))
}

func TestQueryTokenSecurity(t *testing.T) {
	lr.Init()
	Init()
	tokenAddress := "3Q6KfoGoa3zZ65bPcwND4XW2oBxGqisPhCmLSHzQpump"

	res, err := QueryTokenSecurity(tokenAddress, "solana")
	assert.NoError(t, err)
	t.Log(utils.ToJson(res))
}

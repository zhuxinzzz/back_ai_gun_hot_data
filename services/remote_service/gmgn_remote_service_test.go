package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
=== RUN   TestQueryTokensByName

	gmgn_remote_service_test.go:18:
	    [
	      {
	        "name": "HOKK Finance",
	        "symbol": "hokk",
	        "address": "0xe87e15b9c7d989474cb6d8c56b3db4efad5b21e8",
	        "network": "Ethereum",
	        "is_internal": false,
	        "logo": "https://coin-images.coingecko.com/coins/images/14985/small/hokk.png?1696514647",
	        "market_cap": "489547.0",
	        "price_usd": "0.00017063",
	        "decimals": 18
	      },
	      {
	        "name": "Hokkaidu Inu",
	        "symbol": "$hokk",
	        "address": "0xc40af1e4fecfa05ce6bab79dcd8b373d2e436c4e",
	        "network": "Ethereum",
	        "is_internal": false,
	        "logo": "https://coin-images.coingecko.com/coins/images/34890/small/IMG29529FNM3.png?1710943740",
	        "market_cap": "426256.0",
	        "price_usd": "4.304e-12",
	        "decimals": 9
	      },
	      {
	        "name": "Hokkaidu Inu",
	        "symbol": "HOKK",
	        "address": "0x4f2cef6f39114ade3d8af4020fa1de1d064cadaf",
	        "network": "Ethereum",
	        "is_internal": false,
	        "logo": "https://s2.coinmarketcap.com/static/img/coins/64x64/38177.png",
	        "market_cap": "0",
	        "price_usd": "0.0018398597304606255",
	        "decimals": 18
	      },
	      {
	        "name": "Hokkaido Inu Token",
	        "symbol": "hinu",
	        "address": "0x0113c07b3b8e4f41b62d713b5b12616bf2856585",
	        "network": "Ethereum",
	        "is_internal": false,
	        "logo": "https://coin-images.coingecko.com/coins/images/38812/small/1000248760.jpg?1719070391",
	        "market_cap": "0",
	        "price_usd": "2.451e-08",
	        "decimals": 9
	      }
	    ]

--- PASS: TestQueryTokensByName (0.39s)
*/
func TestQueryTokensByName(t *testing.T) {
	lr.Init()
	Init()
	name := "Kraken Wrapped BTC"
	chain := ""

	res, err := QueryTokensByName(name, chain)
	assert.NoError(t, err)
	t.Log(utils.ToJson(res))
}

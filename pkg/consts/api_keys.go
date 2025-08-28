package consts

// API密钥配置
const (
	//Freecryptoapi
	FREE_CRYPTO_API_KEY = "3f4tetz5t3a6kxethg1w"

	// coin layer
	COIN_LAYER_API_KEY = "a3b0b9b07948802474e93f1b67a4570b"

	// CoinAPI.io API密钥 (需要申请)
	COIN_API_KEY = "07b143be-925c-4458-bfe8-27aa32341088" // 请替换为实际的API密钥

	// CoinMarketCap API密钥
	COIN_MARKET_CAP_KEY = "276ceaa5-0cb4-403d-a607-a25f0c544198"

	// Etherscan API密钥
	ETHERSCAN_API_KEY = "5T54C1KY433G322VAY3JF59E12F23UVW9F"

	// CryptoCompare API密钥 (可选)
	CRYPTOCOMPARE_API_KEY = "16299a8751a642682c77d5072f48371c6b99e7ad3f02adc4dda772c03e4692c3"

	// CoinGecko API密钥 (基础版免费)
	COINGECKO_API_KEY = ""

	// CoinRanking API密钥
	COINRANKING_API_KEY = "coinrankingc86beff9815128a3a1d9bf095c95650348e06c6542cada65"

	// CoinCap API密钥
	COINCAP_API_KEY = "3d7e70ad1e964919b462b0247f2a9aa146faa0f91b45eaea2536983f5a5182f8"
)

// API配置
const (
	// CoinGecko API配置
	COINGECKO_BASE_URL        = "https://api.coingecko.com/api/v3"
	COINGECKO_FREE_TIER_LIMIT = 50 // 免费版每分钟请求限制

	// CoinAPI.io API配置
	COINAPI_BASE_URL        = "https://rest.coinapi.io/v1"
	COINAPI_FREE_TIER_LIMIT = 100 // 免费版每天请求限制

	// FreeCryptoAPI配置
	FREE_CRYPTO_API_BASE_URL        = "https://api.freecryptoapi.com/v1"
	FREE_CRYPTO_API_FREE_TIER_LIMIT = 100000 // 免费版每月请求限制

	// CoinLayer配置
	COIN_LAYER_BASE_URL        = "http://api.coinlayer.com"
	COIN_LAYER_FREE_TIER_LIMIT = 100 // 免费版每月请求限制

	// CryptoCompare配置
	CRYPTOCOMPARE_BASE_URL        = "https://min-api.cryptocompare.com/data"
	CRYPTOCOMPARE_FREE_TIER_LIMIT = 100000 // 免费版每月请求限制

	// CoinCap配置
	COINCAP_BASE_URL        = "https://api.coincap.io"
	COINCAP_FREE_TIER_LIMIT = 0 // 完全免费，无限制

	// CoinRanking配置
	COINRANKING_BASE_URL        = "https://coinranking.com/api/v2"
	COINRANKING_FREE_TIER_LIMIT = 0 // 完全免费，无限制

	// CoinMarketCap API配置
	COINMARKETCAP_BASE_URL        = "https://pro-api.coinmarketcap.com"
	COINMARKETCAP_FREE_TIER_LIMIT = 10000 // 免费版每月请求限制
)

// API状态检查
func IsCoinGeckoAvailable() bool {
	return COIN_GECKO_KEY != ""
}

func IsCoinAPIAvailable() bool {
	return COIN_API_KEY != "" && COIN_API_KEY != "YOUR_COINAPI_KEY"
}

func IsFreeCryptoAPIAvailable() bool {
	return FREE_CRYPTO_API_KEY != ""
}

func IsCoinLayerAvailable() bool {
	return COIN_LAYER_API_KEY != ""
}

func IsCryptoCompareAvailable() bool {
	return CRYPTOCOMPARE_API_KEY != ""
}

func IsCoinCapAvailable() bool {
	return COINCAP_API_KEY != ""
}

func IsCoinRankingAvailable() bool {
	return COINRANKING_API_KEY != ""
}

func IsCoinMarketCapAvailable() bool {
	return COIN_MARKET_CAP_KEY != ""
}

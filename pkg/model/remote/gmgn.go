package remote

import (
	"strconv"
	"strings"

	"back_ai_gun_data/pkg/model/dto"
)

// GmGnToken GMGN 代币信息
type GmGnToken struct {
	Address     string `json:"address"`
	Chain       string `json:"chain"`
	ChainID     int    `json:"chain_id"`
	Decimals    int    `json:"decimals"`
	Logo        string `json:"logo"`
	MarketCap   string `json:"market_cap"`
	Name        string `json:"name"`
	Network     string `json:"network"`
	PriceUSD    string `json:"price_usd"`
	Symbol      string `json:"symbol"`
	TotalSupply string `json:"total_supply"`
	Volume24h   string `json:"volume_24h"`
	IsInternal  bool   `json:"is_internal"`
	Liquidity   string `json:"liquidity"`
}

// IsSupportedChain 检查是否为支持的链
func (t *GmGnToken) IsSupportedChain() bool {
	supportedChains := map[string]bool{
		"ethereum":  true,
		"bsc":       true,
		"polygon":   true,
		"arbitrum":  true,
		"optimism":  true,
		"avalanche": true,
		"solana":    true,
		"sui":       true,
	}
	return supportedChains[strings.ToLower(t.Network)]
}

// TokenQueryParams 代币查询参数
type TokenQueryParams struct {
	Q      string `json:"q"`      // 查询关键词
	Chain  string `json:"chain"`  // 链名称
	Limit  int    `json:"limit"`  // 限制数量
	Fuzzy  int    `json:"fuzzy"`  // 是否模糊匹配
	Offset int    `json:"offset"` // 偏移量
}

// TokenQueryResponse 代币查询响应
type TokenQueryResponse struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string][]GmGnToken `json:"data"`
}

// TokenSecurityResp 代币安全信息响应
type TokenSecurityResp struct {
	BalanceMutableAuthority struct { // 余额修改权限
		//Authority []any `json:"authority"`
		Status string `json:"status"` // 0: 已禁用
	} `json:"balance_mutable_authority"`
	Closable struct { // 账户关闭权限
		//Authority []any `json:"authority"`
		Status string `json:"status"` // 0: 已禁用
	} `json:"closable"`
	//Creators                      []any    `json:"creators"`
	DefaultAccountState           string   `json:"default_account_state"` // 默认账户状态 1 已激活
	DefaultAccountStateUpgradable struct { // 默认账户状态升级权限
		// Authority []any `json:"authority"`
		Status string `json:"status"` // 0: 已禁用
	} `json:"default_account_state_upgradable"`
	//Dex       any      `json:"dex"`
	Freezable struct { // 冻结权限
		// Authority []any `json:"authority"`
		Status string `json:"status"` // 0: 已禁用
	} `json:"freezable"`
	HolderCount string     `json:"holder_count"` // 持有者数量
	Holders     []struct { // 持有者列表
		Account  string `json:"account"`
		Balance  string `json:"balance"`
		IsLocked int    `json:"is_locked"`
		// LockedDetail []any `json:"locked_detail"`
		Percent      string `json:"percent"`
		Tag          string `json:"tag"`
		TokenAccount string `json:"token_account"`
	} `json:"holders"`
	//LpHolders any `json:"lp_holders"`
	Metadata struct {
		Description string `json:"description"`
		Name        string `json:"name"`
		Symbol      string `json:"symbol"`
		Uri         string `json:"uri"`
	} `json:"metadata"`
	MetadataMutable struct {
		//MetadataUpgradeAuthority []any  `json:"metadata_upgrade_authority"`
		Status string `json:"status"`
	} `json:"metadata_mutable"`
	Mintable struct {
		// Authority []any `json:"authority"`
		Status string `json:"status"`
	} `json:"mintable"`
	NonTransferable string `json:"non_transferable"`
	TotalSupply     string `json:"total_supply"`
	//TransferFee     struct {
	//} `json:"transfer_fee"`
	TransferFeeUpgradable struct { // 转账费率修改权限
		// Authority []any `json:"authority"`
		Status string `json:"status"` // 0: 已禁用
	} `json:"transfer_fee_upgradable"`
	// TransferHook           []any `json:"transfer_hook"`
	TransferHookUpgradable struct { // 转账钩子升级权限
		// Authority []any `json:"authority"`
		Status string `json:"status"` // 0: 已禁用
	} `json:"transfer_hook_upgradable"`
	TrustedToken int `json:"trusted_token"`
}

// GetRiskLevel 评估代币风险等级
// 返回值: low(低风险，可以买入), middle(中等风险，不推荐买入), high(高风险，不要买入), unknown(未知风险，无法判断)
func (t *TokenSecurityResp) GetRiskLevel() string {
	if t == nil {
		return "unknown"
	}

	// 分层风险评估
	honeypotRisk := t.assessHoneypotRisk()
	rugPullRisk := t.assessRugPullRisk()
	permissionRisk := t.assessPermissionRisk()

	// 综合风险判断
	if honeypotRisk == "high" || rugPullRisk == "high" || permissionRisk == "high" {
		return "high"
	} else if honeypotRisk == "middle" || rugPullRisk == "middle" || permissionRisk == "middle" {
		return "middle"
	} else {
		return "low"
	}
}

// assessHoneypotRisk 评估蜜罐风险
// 蜜罐代币通常指那些允许买入但无法卖出的骗局
func (t *TokenSecurityResp) assessHoneypotRisk() string {
	riskCount := 0

	// 核心蜜罐特征检查
	if t.Freezable.Status == "1" {
		riskCount++ // 可冻结账户，典型蜜罐特征
	}
	if t.TransferFeeUpgradable.Status == "1" {
		riskCount++ // 可修改转账费用，可能设置高额卖税
	}
	if t.TransferHookUpgradable.Status == "1" {
		riskCount++ // 可添加转账钩子，可能阻止卖出
	}
	if t.NonTransferable == "1" {
		riskCount++ // 完全不可转让，明确的高风险
	}

	// 风险等级判断
	if riskCount > 0 {
		return "high"
	}
	return "low"
}

// assessRugPullRisk 评估 Rug Pull 风险
// Rug Pull 指项目方撤走流动性，使代币变得一文不值
func (t *TokenSecurityResp) assessRugPullRisk() string {
	if len(t.Holders) == 0 {
		return "unknown" // 无法获取持有者信息
	}

	// 计算最大持有者占比
	maxPercent := 0.0
	for _, holder := range t.Holders {
		if holder.Percent != "" {
			if percent, err := strconv.ParseFloat(holder.Percent, 64); err == nil {
				if percent > maxPercent {
					maxPercent = percent
				}
			}
		}
	}

	// 持有者数量检查
	holderCount := 0
	if count, err := strconv.Atoi(t.HolderCount); err == nil {
		holderCount = count
	}

	// 风险等级判断
	if maxPercent > 80.0 {
		return "high" // 高度集中，Rug Pull 风险极高
	} else if maxPercent > 50.0 || holderCount <= 5 {
		return "middle" // 中度集中或持有者过少
	}
	return "low"
}

// assessPermissionRisk 评估权限风险
// 检查合约权限是否过于集中，可能被恶意利用
func (t *TokenSecurityResp) assessPermissionRisk() string {
	riskCount := 0

	// 高风险权限检查
	if t.BalanceMutableAuthority.Status == "1" {
		riskCount++ // 可修改余额，可能清零用户代币
	}
	if t.Closable.Status == "1" {
		riskCount++ // 可关闭账户，阻止用户操作
	}
	if t.Mintable.Status == "1" {
		riskCount++ // 可增发代币，稀释价值
	}

	// 中等风险权限检查
	if t.MetadataMutable.Status == "1" {
		riskCount++ // 可修改元数据，可能误导用户
	}

	// 风险等级判断
	if riskCount >= 2 {
		return "high" // 多个高风险权限
	} else if riskCount == 1 {
		return "middle" // 单个风险权限
	}
	return "low"
}

func (t *GmGnToken) ToProjectChainData(chainID string) *dto.ProjectChainData {
	// 解析市值
	var marketCap *float64
	if t.MarketCap != "" {
		if parsed, err := strconv.ParseFloat(t.MarketCap, 64); err == nil {
			marketCap = &parsed
		}
	}

	// 解析价格
	var price *float64
	if t.PriceUSD != "" {
		if parsed, err := strconv.ParseFloat(t.PriceUSD, 64); err == nil {
			price = &parsed
		}
	}

	// 解析24小时交易量
	var volume24h *float64
	if t.Volume24h != "" {
		if parsed, err := strconv.ParseFloat(t.Volume24h, 64); err == nil {
			volume24h = &parsed
		}
	}

	// 设置标准为ERC20（根据传入参数）
	//standard := "ERC20"

	return &dto.ProjectChainData{
		ChainID:         &chainID,
		ContractAddress: t.Address,
		Type:            nil, // 可以根据需要设置
		//Standard:             &standard,
		Decimals:             &t.Decimals,
		Version:              nil, // 可以根据需要设置
		Name:                 &t.Name,
		Symbol:               &t.Symbol,
		Logo:                 &t.Logo,
		LifiCoinKey:          nil, // 可以根据需要设置
		TradingVolume24Hours: volume24h,
		MarketCap24Hours:     marketCap,
		Price24Hours:         price,
		Description:          "", // 可以根据需要设置
		IsVisible:            true,
		IsDeleted:            false,
	}
}

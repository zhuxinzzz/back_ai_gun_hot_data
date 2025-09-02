package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"back_ai_gun_data/pkg/model/remote"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

const (
	queryTokensURL   = "/api/v1/ai/tokens"
	tokenSecurityURL = "/api/v1/ai/tokens/%s/security"
)

func QueryTokens(ctx context.Context, params remote.TokenQueryParams) (*remote.TokenQueryResponse, error) {
	queryParams := url.Values{}
	if params.Q != "" {
		queryParams.Set("q", params.Q)
	}
	if params.Chain != "" {
		queryParams.Set("chain", params.Chain)
	}
	if params.Limit > 0 {
		queryParams.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Fuzzy >= 0 {
		queryParams.Set("fuzzy", strconv.Itoa(params.Fuzzy))
	}

	apiURL := GetHost() + queryTokensURL
	// 手动构建查询字符串，保留空格
	if len(queryParams) > 0 {
		pairs := make([]string, 0, len(queryParams))
		for key, values := range queryParams {
			for _, value := range values {
				// 使用 url.QueryEscape 保留空格为 %20
				escapedKey := url.QueryEscape(key)
				escapedValue := url.QueryEscape(value)
				pairs = append(pairs, escapedKey+"="+escapedValue)
			}
		}
		apiURL += "?" + strings.Join(pairs, "&")
	}

	// 发送请求
	resp, err := Cli().R().
		SetContext(ctx).
		SetResult(&remote.TokenQueryResponse{}).
		Get(apiURL)

	if err != nil {
		lr.E().Error("QueryTokens failed: ", err)
		return nil, fmt.Errorf("query tokens failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("http error: %d", resp.StatusCode())
	}

	result := resp.Result().(*remote.TokenQueryResponse)

	// 检查业务错误
	if result.Code != 0 {
		return nil, fmt.Errorf("business error: %d - %s", result.Code, result.Message)
	}

	return result, nil
}

func QueryTokensByName(name string, chain string) ([]remote.GmGnToken, error) {
	return QueryTokensByNameWithLimit(nil, name, chain, 10) // 默认10，保持向后兼容
}

// SOL→ BSC→ ETH→ Base
var chainFilter = map[string]struct{}{
	"solana":   {},
	"bsc":      {},
	"ethereum": {},
	"eth":      {},
	"base":     {},
}

func QueryTokensByNameWithLimit(ctx context.Context, name string, chain string, limit int) ([]remote.GmGnToken, error) {
	params := remote.TokenQueryParams{
		Q:     name,  // 查询关键字,全称、简称、地址,多个查询用逗号分隔
		Chain: chain, // 指定链
		Limit: limit, // 由调用者控制数量
		Fuzzy: 1,     // 是否为模糊匹配,1:是,0:否 默认值: 1
	}

	resp, err := QueryTokens(ctx, params)
	if err != nil {
		lr.E().Error("QueryTokensByNameWithLimit failed: ", err)
		return nil, err
	}

	var allTokens []remote.GmGnToken
	for _, tokens := range resp.Data {
		allTokens = append(allTokens, tokens...)
	}

	fillteredTokens := make([]remote.GmGnToken, 0, len(allTokens))
	for _, token := range allTokens {
		if _, exists := chainFilter[token.Network]; exists {
			fillteredTokens = append(fillteredTokens, token)
		}
	}

	return fillteredTokens, nil
}

/*好的，这是对您提供的代币安全信息接口响应（resp）中各个字段的详细解释。这份报告分析了一个名为 "HoloworldAI"（代币符号：HOL）的代币。

### 核心摘要
这个代币的合约权限大部分是**不可变的**，这意味着其核心规则（如总供应量、元数据、转账费用等）无法被修改。这是一个积极的安全信号，因为这降低了项目方恶意修改合约的风险。代币供应量高度集中，这可能是一个风险点。

---

### 详细字段解释

#### 权限管理 (Authority Management)
这些字段说明了谁有权（或是否有可能）修改代币的关键属性。在这个例子中，几乎所有权限都已被放弃（状态为 "0"）。

*   **`balance_mutable_authority`**: **余额修改权限**
*   **`status: "0"`**: 表示权限已禁用。没有人可以将任意账户的代币余额修改为任何数量。
*   **`closable`**: **账户关闭权限**
*   **`status: "0"`**: 表示权限已禁用。代币账户不能被项目方或授权方强制关闭。
*   **`default_account_state_upgradable`**: **默认账户状态升级权限**
*   **`status: "0"`**: 表示权限已禁用。无法修改新创建的代币账户的默认状态（例如，默认是冻结还是解冻）。
*   **`freezable`**: **冻结权限**
*   **`status: "0"`**: 表示权限已禁用。项目方或授权方无法冻结任何持有者的代币账户，用户可以自由转账。
*   **`metadata_mutable`**: **元数据修改权限**
*   **`status: "0"`**: 表示权限已禁用。代币的名称、符号、描述等元数据是**不可变的**，无法被篡改。
*   **`mintable`**: **增发权限**
*   **`status: "0"`**: 表示权限已禁用。这意味着不能再铸造（创建）新的 HOL 代币，总供应量是固定的。
*   **`transfer_fee_upgradable`**: **转账费率修改权限**
*   **`status: "0"`**: 表示权限已禁用。无法引入或更改现有的转账费用结构。
*   **`transfer_hook_upgradable`**: **转账钩子升级权限**
*   **`status: "0"`**: 表示权限已禁用。无法添加或更改在代币转账时触发的额外程序逻辑。

#### 代币基本信息 (Basic Token Information)
*   **`metadata`**: **元数据**
*   **`name`: "HoloworldAI"**: 代币的名称。
*   **`symbol`: "HOL"**: 代币的简称/符号。
*   **`description`: "The AI Engine for Storytelling..."**: 代币或项目的简短描述。
*   **`uri`**: 指向更详细元数据（通常是一个 JSON 文件）的链接，其中可能包含项目官网、Logo图像等信息。
*   **`total_supply`: "1000000000"**: 代币的总供应量为 10 亿枚。
*   **`creators`**: **创建者**
*   此字段为空，通常在 NFT 中会列出创作者地址。
*   **`dex`**: **去中心化交易所信息**
*   `null` 表示该接口未查询到或该代币尚未在主流的去中心化交易所（DEX）上创建流动性池。

#### 持有者与分布 (Holders & Distribution)
*   **`holder_count`: "2"**: 目前有 2 个地址持有该代币。
*   **`holders`**: **持有者列表**
*   **地址 1 (`D5TT...UWd`)**: 持有约 9.999 亿枚代币，占总供应量的 **100.00%**。这是一个高度集中的信号，意味着绝大多数代币由单个地址控制。
*   **地址 2 (`DCpd...dSb`)**: 持有约 0.769 枚代币，占总供应量的 **0.0000%**。
*   **`lp_holders`**: **流动性池（LP）持有者**
*   `null` 表示没有查询到相关的流动性池代币持有者信息。

#### 代币状态与特性 (Token State & Features)
*   **`default_account_state`: "1"**: **默认账户状态**
*   这通常表示新创建的代币账户默认是“已激活”或“未冻结”状态，可以直接进行交易。
*   **`non_transferable`: "0"**: **是否可转让**
*   状态为 "0" 表示该代币是**可以自由转让的**。
*   **`transfer_fee`**: **转账费用**
*   空对象 `{}` 表示目前没有设置转账费用。
*   **`transfer_hook`**: **转账钩子**
*   空数组 `[]` 表示没有设置在转账时会触发的额外智能合约逻辑。
*   **`trusted_token`: 0**: **是否为可信代币**
*   状态为 0 通常表示该代币未经特定平台或机构的“可信”认证。

---

### 安全风险与总结
*   **优点**:
*   **合约权限锁定**: 大部分关键权限（增发、冻结、修改元数据等）都已放弃，降低了项目方作恶的风险。
*   **无交易税**: 目前没有设置转账费用。

*   **风险点**:
*   **高度中心化**: 几乎所有的代币都集中在单个地址中。如果该地址开始大量出售，可能会对代币价格造成巨大冲击。
*   **持有者数量少**: 仅有 2 个持有者，表明该代币的社区参与度和分布非常低，流动性可能很差。*/

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

func QueryTokenSecurity(address string, platform string) (*TokenSecurityResp, error) {
	apiURL := GetHost() + fmt.Sprintf(tokenSecurityURL, address) + "?platform=" + platform
	resp, err := Cli().R().Get(apiURL)
	if err != nil {
		lr.E().Error("QueryTokenSecurity failed: ", err)
		return nil, err
	}

	var result TokenSecurityResp
	dataStr := gjson.Get(resp.String(), "data").Raw
	if err := jsoniter.Unmarshal([]byte(dataStr), &result); err != nil {
		lr.E().Errorf("Failed to unmarshal data: %v", err)
		return nil, err
	}

	return &result, nil
}

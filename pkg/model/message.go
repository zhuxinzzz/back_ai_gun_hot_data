package model

import "encoding/json"

type BaseMessage struct {
	ID        string `json:"id"`        // 情报 ID
	Timestamp int64  `json:"timestamp"` // 时间戳
	Version   string `json:"version"`   // 版本号
}

// IntelligenceMessage 情报排序队列消息结构
type IntelligenceMessage struct {
	BaseMessage
	Data IntelligenceData `json:"data"`
}

// IntelligenceData 情报数据
type IntelligenceData struct {
	IsValuable   bool                 `json:"is_valuable"`
	IsVisible    bool                 `json:"is_visible"`
	AnalyzedTime int64                `json:"analyzed_time"`
	Analyzed     map[string]string    `json:"analyzed"`
	CreatedAt    string               `json:"created_at"`
	Abstract     string               `json:"abstract"`
	Type         string               `json:"type"`
	Title        string               `json:"title"`
	ExtraDatas   ExtraDatas           `json:"extra_datas"`
	Content      string               `json:"content"`
	SourceURL    string               `json:"source_url"`
	Tags         []string             `json:"tags"`
	Score        float64              `json:"score"`
	Medias       []interface{}        `json:"medias"`
	IsDeleted    bool                 `json:"is_deleted"`
	UpdatedAt    string               `json:"updated_at"`
	Subtype      string               `json:"subtype"`
	Entities     []IntelligenceEntity `json:"entities"`
	ID           string               `json:"id"`
	SourceID     string               `json:"source_id"`
	PublishedAt  string               `json:"published_at"`
}

// ExtraDatas 额外数据
type ExtraDatas struct {
	URLs         []interface{} `json:"urls"`
	Quote        interface{}   `json:"quote"`
	Threads      []interface{} `json:"threads"`
	OutsideData  []interface{} `json:"outside_data"`
	UserMentions []interface{} `json:"user_mentions"`
	Article      []interface{} `json:"article"`
	Repost       interface{}   `json:"repost"`
}

// IntelligenceEntity 情报实体
type IntelligenceEntity struct {
	Standard        *string `json:"standard"`
	PriceUSD        string  `json:"price_usd"`
	Symbol          string  `json:"symbol"`
	IsVerified      bool    `json:"isVerified"`
	ContractAddress string  `json:"contractAddress"`
	EntityID        string  `json:"entityId"`
	IsVisible       bool    `json:"isVisible"`
	Type            string  `json:"type"`
	Version         string  `json:"version"`
	CreatedAt       string  `json:"createdAt"`
	MarketCap       string  `json:"market_cap"`
	IsDeleted       bool    `json:"isDeleted"`
	ChainID         string  `json:"chainId"`
	Decimals        int     `json:"decimals"`
	Name            string  `json:"name"`
	Logo            *string `json:"logo"`
	ID              string  `json:"id"`
	ProjectID       *string `json:"projectId"`
	UpdatedAt       string  `json:"updatedAt"`
}

// ETLEntityMessage ETL实体数据队列消息结构
type ETLEntityMessage struct {
	BaseMessage
	Data            ETLEntityData          `json:"data"`
	ProfileTime     json.Number            `json:"profile_time"`
	Subtype         string                 `json:"subtype"`
	Analyzed        map[string]interface{} `json:"analyzed"`
	Type            string                 `json:"type"`
	AnalyzedTime    int64                  `json:"analyzed_time"`
	EntitiesExtract EntitiesExtract        `json:"entities_extract"`
	Investment      Investment             `json:"investment"`
}

// ETLEntityData ETL实体数据
type ETLEntityData struct {
	Comments    []Comment     `json:"comments"`
	SenderInfo  SenderInfo    `json:"sender_info"`
	TweetID     int64         `json:"tweet_id"`
	Threads     []interface{} `json:"threads"`
	Abstract    string        `json:"abstract"`
	Title       string        `json:"title"`
	Article     []interface{} `json:"article"`
	Content     string        `json:"content"`
	SourceURL   string        `json:"source_url"`
	Tags        []interface{} `json:"tags"`
	Medias      []Media       `json:"medias"`
	URLs        []interface{} `json:"urls"`
	ID          string        `json:"id"`
	SourceID    string        `json:"source_id"`
	OutsideData []interface{} `json:"outside_data"`
	PublishedAt int64         `json:"published_at"`
}

// Comment 评论信息
type Comment struct {
	SenderInfo  SenderInfo    `json:"sender_info"`
	TweetID     int64         `json:"tweet_id"`
	Threads     []interface{} `json:"threads"`
	Abstract    string        `json:"abstract"`
	Title       string        `json:"title"`
	Article     []interface{} `json:"article"`
	Content     string        `json:"content"`
	SourceURL   string        `json:"source_url"`
	Tags        []interface{} `json:"tags"`
	Medias      []Media       `json:"medias"`
	URLs        []interface{} `json:"urls"`
	OutsideData []interface{} `json:"outside_data"`
	PublishedAt int64         `json:"published_at"`
}

// Media 媒体信息
type Media struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// EntitiesExtract 实体提取结果
type EntitiesExtract struct {
	Entities struct {
		Persons   []string `json:"persons"`
		Addresses []string `json:"addresses"`
		Projects  []string `json:"projects"`
		Tokens    []string `json:"tokens"`
		Accounts  []string `json:"accounts"`
		Token     []string `json:"token"`
	} `json:"entities"`
	Timings int `json:"timings"`
}

// Investment 投资分析
type Investment struct {
	Score      float64  `json:"score"`
	Reason     string   `json:"reason"`
	Entities   []string `json:"entities"`
	IsValuable bool     `json:"is_valuable"`
}

// Token 代币信息
type Token struct {
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	ContractAddress string `json:"contract_address"`
	Network         string `json:"network"`
	Decimals        int    `json:"decimals"`
	Logo            string `json:"logo"`
}

// SenderInfo 发送者信息
type SenderInfo struct {
	VerifiedStatus     string        `json:"verified_status"`
	DisplayURLs        []URLInfo     `json:"display_urls"`
	Banner             string        `json:"banner"`
	Description        string        `json:"description"`
	Avatar             string        `json:"avatar"`
	FollowerCount      int           `json:"follower_count"`
	JoinedAt           int64         `json:"joined_at"`
	DescURLs           []URLInfo     `json:"desc_urls"`
	ScreenName         string        `json:"screen_name"`
	FollowingCount     int           `json:"following_count"`
	Name               string        `json:"name"`
	Location           string        `json:"location"`
	Categories         []interface{} `json:"categories"`
	TwitterID          int64         `json:"twitter_id"`
	AffHighlightLabels []interface{} `json:"aff_highlight_labels"`
}

// URLInfo URL信息
type URLInfo struct {
	DisplayURL  string `json:"display_url"`
	ExpandedURL string `json:"expanded_url"`
	URL         string `json:"url"`
}

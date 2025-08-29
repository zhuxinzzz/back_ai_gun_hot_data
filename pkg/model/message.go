package model

type MessageData struct {
	ID        string    `json:"id"` // 情报 id
	Timestamp int64     `json:"timestamp"`
	Version   string    `json:"version"`
	Data      TweetData `json:"data"`
}

type TweetData struct {
	SenderInfo      SenderInfo `json:"sender_info"`
	TweetID         string     `json:"tweet_id"`
	Content         string     `json:"content"`
	SourceURL       string     `json:"source_url"`
	PublishedAt     int64      `json:"published_at"`
	Type            string     `json:"type"`
	Subtype         string     `json:"subtype"`
	AnalyzedTime    int64      `json:"analyzed_time"`
	EntitiesExtract struct {
		Entities struct {
			Tokens   []string `json:"tokens"` // token 名字集合，一个情报可能涉及多个 token
			Projects []string `json:"projects"`
			Persons  []string `json:"persons"`
			Accounts []string `json:"accounts"`
		} `json:"entities"`
		Timings int `json:"timings"`
	} `json:"entities_extract"`
}

type SenderInfo struct {
	ScreenName     string `json:"screen_name"`
	Name           string `json:"name"`
	FollowerCount  int    `json:"follower_count"`
	FollowingCount int    `json:"following_count"`
	Description    string `json:"description"`
	Location       string `json:"location"`
	Avatar         string `json:"avatar"`
}

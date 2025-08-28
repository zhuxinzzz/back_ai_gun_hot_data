package service

import (
	"back_ai_gun_data/pkg/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/lr"
)

// 处理消息数据 - 主业务函数
func ProcessMessageData(data *model.MessageData) error {
	// 验证消息数据
	if err := validateMessage(data); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	// 记录处理开始
	lr.I().Infof("Processing message: %s, tweet: %s", data.ID, data.Data.TweetID)

	// 提取关键信息
	tweetInfo := extractTweetInfo(data)

	// 分析实体
	entities := analyzeEntities(data)

	// 存储数据
	if err := storeMessageData(data, tweetInfo, entities); err != nil {
		return fmt.Errorf("failed to store message data: %w", err)
	}

	// 记录处理完成
	lr.I().Infof("Message processed successfully: %s", data.ID)

	return nil
}

// 验证消息数据
func validateMessage(data *model.MessageData) error {
	if data == nil {
		return fmt.Errorf("message data is nil")
	}

	if data.ID == "" {
		return fmt.Errorf("message ID is required")
	}

	if data.Data.TweetID == "" {
		return fmt.Errorf("tweet ID is required")
	}

	if data.Data.SenderInfo.ScreenName == "" {
		return fmt.Errorf("sender screen name is required")
	}

	return nil
}

// 提取推文信息
func extractTweetInfo(data *model.MessageData) map[string]interface{} {
	info := map[string]interface{}{
		"tweet_id":      data.Data.TweetID,
		"content":       data.Data.Content,
		"source_url":    data.Data.SourceURL,
		"published_at":  time.Unix(data.Data.PublishedAt/1000, 0).Format(time.RFC3339),
		"type":          data.Data.Type,
		"subtype":       data.Data.Subtype,
		"analyzed_time": time.Unix(data.Data.AnalyzedTime/1000, 0).Format(time.RFC3339),
	}

	// 添加发送者信息
	if data.Data.SenderInfo.ScreenName != "" {
		info["sender"] = map[string]interface{}{
			"screen_name":     data.Data.SenderInfo.ScreenName,
			"name":            data.Data.SenderInfo.Name,
			"follower_count":  data.Data.SenderInfo.FollowerCount,
			"following_count": data.Data.SenderInfo.FollowingCount,
			"description":     data.Data.SenderInfo.Description,
			"location":        data.Data.SenderInfo.Location,
			"avatar":          data.Data.SenderInfo.Avatar,
		}
	}

	return info
}

// 分析实体信息
func analyzeEntities(data *model.MessageData) map[string]interface{} {
	entities := data.Data.EntitiesExtract.Entities

	result := map[string]interface{}{
		"tokens":   entities.Tokens,
		"projects": entities.Projects,
		"persons":  entities.Persons,
		"accounts": entities.Accounts,
	}

	// 记录实体分析结果
	if len(entities.Tokens) > 0 {
		lr.I().Infof("Extracted tokens: %v", entities.Tokens)
	}

	if len(entities.Persons) > 0 {
		lr.I().Infof("Extracted persons: %v", entities.Persons)
	}

	if len(entities.Accounts) > 0 {
		lr.I().Infof("Extracted accounts: %v", entities.Accounts)
	}

	return result
}

// 存储消息数据
func storeMessageData(data *model.MessageData, tweetInfo map[string]interface{}, entities map[string]interface{}) error {
	// 这里应该实现实际的数据存储逻辑
	// 例如：存储到数据库、写入文件、发送到其他服务等

	// 示例：将数据转换为JSON并记录
	storageData := map[string]interface{}{
		"message_id": data.ID,
		"timestamp":  data.Timestamp,
		"version":    data.Version,
		"tweet_info": tweetInfo,
		"entities":   entities,
		"stored_at":  time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(storageData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal storage data: %w", err)
	}

	// 使用缓存存储数据
	ctx := context.Background()
	cacheKey := fmt.Sprintf("message:%s", data.ID)

	// 存储到Redis缓存，过期时间30分钟
	if err := cache.Set(ctx, cacheKey, string(jsonData), 30*time.Minute); err != nil {
		lr.E().Errorf("Failed to cache message data: %v", err)
		// 缓存失败不影响主流程，只记录错误
	}

	lr.I().Infof("Stored data for message %s:\n%s", data.ID, string(jsonData))

	return nil
}

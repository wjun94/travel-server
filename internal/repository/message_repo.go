package repository

import (
	"time"

	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// CreateMessage 发送消息
func CreateMessage(msg *model.Message) error {
	return database.DB.Create(msg).Error
}

// GetMessagesBetweenUsers 获取两个用户之间的私聊记录
func GetMessagesBetweenUsers(user1, user2 string) ([]model.Message, error) {
	var msgs []model.Message
	err := database.DB.Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		user1, user2, user2, user1).Order("created_at asc").Find(&msgs).Error
	return msgs, err
}

// ConversationItem 会话列表项
type ConversationItem struct {
	UserID      string    `json:"userId"`
	Nickname    string    `json:"nickname"`
	AvatarURL   string    `json:"avatarUrl"`
	LastContent string    `json:"lastContent"`
	LastTime    time.Time `json:"lastTime"`
	UnreadCount int64     `json:"unreadCount"`
}

// GetConversationList 获取当前用户的会话列表（按最后消息时间倒序）
func GetConversationList(userID string) ([]ConversationItem, error) {
	// 1. 查出所有与该用户有关的消息（本人发出或接收），按时间倒序
	type msgRow struct {
		FromUserID string
		ToUserID   string
		Content    string
		CreatedAt  time.Time
	}
	var msgs []msgRow
	database.DB.Model(&model.Message{}).
		Select("from_user_id, to_user_id, content, created_at").
		Where("(from_user_id = ? OR to_user_id = ?) AND type = 1", userID, userID).
		Order("created_at desc").
		Find(&msgs)

	// 2. 按对方用户去重，取每条消息的最新一条作为最后消息
	seen := make(map[string]bool)
	var otherIDs []string
	lastContent := make(map[string]string)
	lastTime := make(map[string]time.Time)
	for _, m := range msgs {
		otherID := m.ToUserID
		if m.ToUserID == userID {
			otherID = m.FromUserID
		}
		if seen[otherID] {
			continue
		}
		seen[otherID] = true
		otherIDs = append(otherIDs, otherID)
		lastContent[otherID] = m.Content
		lastTime[otherID] = m.CreatedAt
	}

	if len(otherIDs) == 0 {
		return []ConversationItem{}, nil
	}

	// 3. 批量查对方用户信息
	var users []model.User
	database.DB.Select("id, nickname, avatar_url").Where("id IN ?", otherIDs).Find(&users)
	userMap := make(map[string]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// 4. 批量查未读消息数（发给当前用户且未读）
	unreadMap := make(map[string]int64)
	for _, oid := range otherIDs {
		var cnt int64
		database.DB.Model(&model.Message{}).
			Where("from_user_id = ? AND to_user_id = ? AND is_read = 0 AND type = 1", oid, userID).
			Count(&cnt)
		unreadMap[oid] = cnt
	}

	// 5. 组装结果
	result := make([]ConversationItem, 0, len(otherIDs))
	for _, oid := range otherIDs {
		u := userMap[oid]
		result = append(result, ConversationItem{
			UserID:      oid,
			Nickname:    u.Nickname,
			AvatarURL:   u.AvatarURL,
			LastContent: lastContent[oid],
			LastTime:    lastTime[oid],
			UnreadCount: unreadMap[oid],
		})
	}
	return result, nil
}

package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// CreateMessage 发送消息
func CreateMessage(msg *model.Message) error {
	return database.DB.Create(msg).Error
}

// MarkMessagesAsRead 标记两个用户之间的私聊消息为已读（from → to 的消息）
func MarkMessagesAsRead(fromUserID, toUserID string) error {
	return database.DB.Model(&model.Message{}).
		Where("from_user_id = ? AND to_user_id = ? AND type = 1 AND is_read = 0", fromUserID, toUserID).
		Update("is_read", 1).Error
}

// GetMessagesBetweenUsers 分页获取两个用户之间的私聊记录（最新一页，按时间正序返回）
func GetMessagesBetweenUsers(user1, user2 string, page, pageSize int) ([]model.Message, int64, error) {
	var msgs []model.Message
	var total int64
	query := database.DB.Model(&model.Message{}).
		Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
			user1, user2, user2, user1)
	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&msgs).Error
	// 倒序取页后反转成正序，保证聊天记录从旧到新展示
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, total, err
}

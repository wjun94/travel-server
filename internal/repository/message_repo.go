package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// CreateMessage 发送消息
func CreateMessage(msg *model.Message) error {
	return database.DB.Create(msg).Error
}

// GetMessagesBetweenUsers 获取两个用户之间的私聊记录
func GetMessagesBetweenUsers(user1, user2 uint) ([]model.Message, error) {
	var msgs []model.Message
	err := database.DB.Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		user1, user2, user2, user1).Order("created_at asc").Find(&msgs).Error
	return msgs, err
}

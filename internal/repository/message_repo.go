package repository

import (
	"time"

	"gorm.io/gorm"

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
// 已过滤当前用户删除的消息：发送者删除的只看接收者视角，接收者删除的只看发送者视角
func GetMessagesBetweenUsers(user1, user2 string, page, pageSize int) ([]model.Message, int64, error) {
	var msgs []model.Message
	var total int64
	query := database.DB.Model(&model.Message{}).
		Where("(from_user_id = ? AND to_user_id = ? AND deleted_by_sender = 0) OR (from_user_id = ? AND to_user_id = ? AND deleted_by_receiver = 0)",
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

// UpsertChatSession 创建或恢复私聊会话（软删记录自动恢复，用于对方重新发消息时会话重新出现）
func UpsertChatSession(userID, peerID string) error {
	var session model.ChatSession
	err := database.DB.Unscoped().Where("user_id = ? AND peer_id = ?", userID, peerID).First(&session).Error
	if err == gorm.ErrRecordNotFound {
		return database.DB.Create(&model.ChatSession{
			UserID: userID,
			PeerID: peerID,
		}).Error
	}
	if err != nil {
		return err
	}
	// 已存在：恢复软删并刷新时间
	return database.DB.Unscoped().Model(&model.ChatSession{}).
		Where("id = ?", session.ID).
		Updates(map[string]interface{}{
			"deleted_at": nil,
			"updated_at": time.Now(),
		}).Error
}

// ClearChatHistory 清空与某用户的聊天记录（对当前用户隐藏全部消息，会话保留）
func ClearChatHistory(userID, peerID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 我发出的消息：发送者侧删除
		if err := tx.Model(&model.Message{}).
			Where("from_user_id = ? AND to_user_id = ? AND type = 1", userID, peerID).
			Update("deleted_by_sender", 1).Error; err != nil {
			return err
		}
		// 对方发给我的消息：接收者侧删除
		return tx.Model(&model.Message{}).
			Where("from_user_id = ? AND to_user_id = ? AND type = 1", peerID, userID).
			Update("deleted_by_receiver", 1).Error
	})
}

// DeleteChatSession 删除私聊会话（清空聊天记录 + 会话从列表消失）
func DeleteChatSession(userID, peerID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 清空聊天记录
		if err := tx.Model(&model.Message{}).
			Where("from_user_id = ? AND to_user_id = ? AND type = 1", userID, peerID).
			Update("deleted_by_sender", 1).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Message{}).
			Where("from_user_id = ? AND to_user_id = ? AND type = 1", peerID, userID).
			Update("deleted_by_receiver", 1).Error; err != nil {
			return err
		}
		// 软删会话
		return tx.Where("user_id = ? AND peer_id = ?", userID, peerID).
			Delete(&model.ChatSession{}).Error
	})
}

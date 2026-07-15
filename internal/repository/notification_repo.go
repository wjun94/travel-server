package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// CreateNotification 创建通知
func CreateNotification(n *model.Notification) error {
	return database.DB.Create(n).Error
}

// MarkNotificationRead 标记单条通知为已读（需校验归属）
func MarkNotificationRead(id, userID string) error {
	return database.DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", 1).Error
}

// MarkAllNotificationsRead 标记当前用户所有通知为已读
func MarkAllNotificationsRead(userID string) error {
	return database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = 0", userID).
		Update("is_read", 1).Error
}

// GetUnreadCounts 获取当前用户各类未读通知的数量
func GetUnreadCounts(userID string) (partnerApplyCount, likeCount, followCount, systemNotifyCount int64, err error) {
	// 1. 搭子申请：我的搭子的待审核申请数
	var partnerIDs []string
	database.DB.Model(&model.Partner{}).Select("id").
		Where("user_id = ?", userID).Find(&partnerIDs)
	if len(partnerIDs) > 0 {
		database.DB.Model(&model.PartnerApplication{}).
			Where("partner_id IN ? AND status = 0", partnerIDs).
			Count(&partnerApplyCount)
	}

	// 2. 点赞通知
	database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND type = 2 AND is_read = 0", userID).
		Count(&likeCount)

	// 3. 新增关注通知
	database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND type = 3 AND is_read = 0", userID).
		Count(&followCount)

	// 4. 系统通知
	database.DB.Model(&model.Message{}).
		Where("to_user_id = ? AND type = 2 AND is_read = 0", userID).
		Count(&systemNotifyCount)

	return
}

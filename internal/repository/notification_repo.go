package repository

import (
	"errors"

	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// GetNotificationByID 获取单条通知（校验归属，仅本人可见）
func GetNotificationByID(id, userID string) (*model.Notification, error) {
	var n model.Notification
	err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&n).Error
	return &n, err
}

// CreateNotification 创建通知
func CreateNotification(n *model.Notification) error {
	return database.DB.Create(n).Error
}

// ListNotifications 分页获取通知列表，type=0 表示全部
func ListNotifications(userID string, notiType, page, pageSize int) ([]model.Notification, int64, error) {
	var list []model.Notification
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Notification{}).Where("user_id = ?", userID)
	if notiType == 1 {
		// 搭子申请分类包含搭子动态（type=6 解散/退出/补位）
		query = query.Where("type IN (1, 6)")
	} else if notiType > 0 {
		query = query.Where("type = ?", notiType)
	}
	query.Count(&total)
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// MarkNotificationRead 标记单条通知为已读/未读（需校验归属；幂等：重复标记不报错）
func MarkNotificationRead(id, userID string, isRead int) error {
	// 先校验通知存在且归属本人，避免依赖 UPDATE 影响行数判断（重复标记已读时影响行数为 0）
	var n model.Notification
	if err := database.DB.Select("id").Where("id = ? AND user_id = ?", id, userID).First(&n).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("通知不存在或无权操作")
		}
		return err
	}
	return database.DB.Model(&model.Notification{}).Where("id = ?", id).Update("is_read", isRead).Error
}

// MarkAllNotificationsRead 标记当前用户所有通知为已读
func MarkAllNotificationsRead(userID string) error {
	return database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = 0", userID).
		Update("is_read", 1).Error
}

// MarkTypeNotificationsRead 标记当前用户指定类型的所有通知为已读（点击tab清空未读数）
func MarkTypeNotificationsRead(userID string, notiType int) error {
	q := database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = 0", userID)
	if notiType == 1 {
		// 搭子申请分类包含搭子动态（type=6 解散/退出/补位）
		q = q.Where("type IN (1, 6)")
	} else {
		q = q.Where("type = ?", notiType)
	}
	return q.Update("is_read", 1).Error
}

// DeleteSystemNotifications 清空当前用户的全部系统通知（type=4）
func DeleteSystemNotifications(userID string) error {
	return database.DB.Where("user_id = ? AND type = 4", userID).
		Delete(&model.Notification{}).Error
}

// GetUnreadCounts 获取当前用户各类未读通知的数量（统一按 Notification 表统计）
func GetUnreadCounts(userID string) (partnerApplyCount, likeCount, followCount, commentCount, systemNotifyCount, partnerDynamicCount int64, err error) {
	// 1. 搭子申请通知（含搭子动态 type=6）
	database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND type IN (1, 6) AND is_read = 0", userID).
		Count(&partnerApplyCount)

	// 2. 点赞通知
	database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND type = 2 AND is_read = 0", userID).
		Count(&likeCount)

	// 3. 新增关注通知
	database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND type = 3 AND is_read = 0", userID).
		Count(&followCount)

	// 4. 评论通知（Notification type=5）
	database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND type = 5 AND is_read = 0", userID).
		Count(&commentCount)

	// 5. 系统通知（Notification type=4）
	database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND type = 4 AND is_read = 0", userID).
		Count(&systemNotifyCount)

	// 6. 搭子动态通知（Notification type=6）
	database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND type = 6 AND is_read = 0", userID).
		Count(&partnerDynamicCount)

	return
}

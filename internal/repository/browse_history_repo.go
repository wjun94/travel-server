package repository

import (
	"time"

	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// AddBrowseHistory 添加浏览记录（相同用户+目标去重，保留最新）
func AddBrowseHistory(bh *model.BrowseHistory) error {
	// 删除同用户同目标的旧记录
	database.DB.Where("user_id = ? AND target_id = ? AND target_type = ?", bh.UserID, bh.TargetID, bh.TargetType).
		Delete(&model.BrowseHistory{})
	// 插入新记录
	return database.DB.Create(bh).Error
}

// GetBrowseHistory 获取用户的浏览历史（分页）
func GetBrowseHistory(userID string, page, pageSize int) ([]model.BrowseHistory, int64, error) {
	var list []model.BrowseHistory
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.BrowseHistory{}).Where("user_id = ?", userID).Count(&total)
	err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// CleanupBrowseHistory 清理超过 retentionDays 的浏览历史
func CleanupBrowseHistory(retentionDays int) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	database.DB.Where("created_at < ?", cutoff).Delete(&model.BrowseHistory{})
}

package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// AddFavorite 添加收藏
func AddFavorite(fav *model.Favorite) error {
	return database.DB.Create(fav).Error
}

// RemoveFavorite 取消收藏
func RemoveFavorite(userID, targetID uint, targetType string) error {
	return database.DB.Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).
		Delete(&model.Favorite{}).Error
}

// IsFavorited 是否已收藏
func IsFavorited(userID, targetID uint, targetType string) bool {
	var count int64
	database.DB.Model(&model.Favorite{}).
		Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).
		Count(&count)
	return count > 0
}

// ListUserFavorites 用户收藏列表
func ListUserFavorites(userID uint, targetType string, page, pageSize int) ([]model.Favorite, int64, error) {
	var favs []model.Favorite
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Favorite{}).Where("user_id = ?", userID)
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	query.Count(&total)
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&favs).Error
	return favs, total, err
}

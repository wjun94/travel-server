package repository

import (
	"errors"

	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// FavoriteItem 收藏项（含目标标题和封面图）
type FavoriteItem struct {
	model.Favorite
	Title      string `json:"title"`
	CoverImage string `json:"coverImage"`
}

// AddFavorite 添加收藏
func AddFavorite(fav *model.Favorite) error {
	// 默认收藏（点赞走独立 Like 逻辑，action=1）
	if fav.Action == 0 {
		fav.Action = 2
	}
	tx := database.DB.Begin()
	if err := tx.Create(fav).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 搭子的收藏与点赞分离，收藏时只同步增加收藏数
	if fav.TargetType == "partner" {
		if err := tx.Model(&model.Partner{}).
			Where("id = ?", fav.TargetID).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// RemoveFavorite 取消收藏
func RemoveFavorite(userID, targetID string, targetType string) error {
	// 先按 targetID + targetType 匹配记录（前端传目标ID）
	var fav model.Favorite
	err := database.DB.Where("user_id = ? AND target_id = ? AND target_type = ? AND action = 2", userID, targetID, targetType).
		First(&fav).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// 兜底：按主键 ID 查找（前端可能传的是 Favorite 记录 ID）
		if err := database.DB.Where("id = ? AND user_id = ?", targetID, userID).
			First(&fav).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("未找到收藏记录")
			}
			return err
		}
	}
	if err := database.DB.Delete(&model.Favorite{}, "id = ?", fav.ID).Error; err != nil {
		return err
	}
	// 搭子的收藏与点赞分离，取消收藏时只同步减少收藏数（下限0）
	if fav.TargetType == "partner" {
		return database.DB.Model(&model.Partner{}).
			Where("id = ?", fav.TargetID).
			UpdateColumn("favorite_count", gorm.Expr("GREATEST(favorite_count - 1, 0)")).Error
	}
	return nil
}

// IsFavorited 是否已收藏
func IsFavorited(userID, targetID string, targetType string) bool {
	var count int64
	database.DB.Model(&model.Favorite{}).
		Where("user_id = ? AND target_id = ? AND target_type = ? AND action = 2", userID, targetID, targetType).
		Count(&count)
	return count > 0
}

// ListUserFavorites 用户收藏列表（含目标标题和封面图）
func ListUserFavorites(userID string, targetType string, page, pageSize int) ([]FavoriteItem, int64, error) {
	var favs []model.Favorite
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Favorite{}).Where("user_id = ? AND action = 2", userID)
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	query.Count(&total)
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&favs).Error
	if err != nil {
		return nil, total, err
	}
	// 组装标题和封面图
	items := make([]FavoriteItem, len(favs))
	for i, f := range favs {
		items[i] = FavoriteItem{Favorite: f}
		switch f.TargetType {
		case "guide":
			var g struct{ Title, CoverImage string }
			database.DB.Model(&model.Guide{}).Select("title", "cover_image").
				Where("id = ?", f.TargetID).Scan(&g)
			items[i].Title = g.Title
			items[i].CoverImage = g.CoverImage
		case "trip":
			var t struct{ Title, CoverImage string }
			database.DB.Model(&model.Trip{}).Select("title", "cover_image").
				Where("id = ?", f.TargetID).Scan(&t)
			items[i].Title = t.Title
			items[i].CoverImage = t.CoverImage
		case "partner":
			var p struct{ Title, Cover string }
			database.DB.Model(&model.Partner{}).Select("title", "cover").
				Where("id = ?", f.TargetID).Scan(&p)
			items[i].Title = p.Title
			items[i].CoverImage = p.Cover
		}
	}
	return items, total, nil
}

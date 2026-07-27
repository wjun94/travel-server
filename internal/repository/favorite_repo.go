package repository

import (
	"errors"

	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// FavoriteItem 收藏项（含目标标题和封面图）
type FavoriteItem struct {
	model.Favorite
	Title      string `json:"title"`
	CoverImage string `json:"coverImage"`
}

// AddFavorite 添加收藏
func AddFavorite(fav *model.Favorite) error {
	return database.DB.Create(fav).Error
}

// RemoveFavorite 取消收藏
func RemoveFavorite(userID, targetID string, targetType string) error {
	// 按 targetID + targetType 匹配删除
	r := database.DB.Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).
		Delete(&model.Favorite{})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected > 0 {
		return nil
	}

	// 兜底：按主键 ID 删除（前端可能传的是 Favorite 记录 ID）
	r = database.DB.Where("id = ? AND user_id = ?", targetID, userID).
		Delete(&model.Favorite{})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected > 0 {
		return nil
	}

	return errors.New("未找到收藏记录")
}

// IsFavorited 是否已收藏
func IsFavorited(userID, targetID string, targetType string) bool {
	var count int64
	database.DB.Model(&model.Favorite{}).
		Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).
		Count(&count)
	return count > 0
}

// ListUserFavorites 用户收藏列表（含目标标题和封面图）
func ListUserFavorites(userID string, targetType string, page, pageSize int) ([]FavoriteItem, int64, error) {
	var favs []model.Favorite
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Favorite{}).Where("user_id = ?", userID)
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
			var t struct{ Title string }
			database.DB.Model(&model.Trip{}).Select("title").
				Where("id = ?", f.TargetID).Scan(&t)
			items[i].Title = t.Title
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

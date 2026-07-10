package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// ==================== Guide 攻略 ====================

// GetGuideFeed 获取已发布的攻略瀑布流（含作者信息、当前用户是否点赞）
func GetGuideFeed(page, pageSize int, destination, userID string) ([]model.GuideFeedItem, int64, error) {
	var guides []model.Guide
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Guide{}).Where("status = ?", 1)
	if destination != "" {
		query = query.Where("destination LIKE ?", "%"+destination+"%")
	}
	query.Count(&total)
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&guides).Error
	if err != nil {
		return nil, 0, err
	}
	// 批量获取作者信息
	userIDs := make([]string, 0, len(guides))
	guideIDs := make([]string, 0, len(guides))
	userIDSet := make(map[string]bool)
	for _, g := range guides {
		if !userIDSet[g.UserID] {
			userIDs = append(userIDs, g.UserID)
			userIDSet[g.UserID] = true
		}
		guideIDs = append(guideIDs, g.ID)
	}
	var users []model.User
	database.DB.Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[string]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}
	// 批量获取当前用户的点赞状态
	likedSet := make(map[string]bool)
	if userID != "" {
		var favs []model.Favorite
		database.DB.Where("user_id = ? AND target_type = ? AND target_id IN ?", userID, "guide", guideIDs).Find(&favs)
		for _, f := range favs {
			likedSet[f.TargetID] = true
		}
	}

	// 批量查询行程天数（按 guide_id 统计不同 day_number）
	type dayCount struct {
		GuideID string
		Days    int
	}
	var dayCounts []dayCount
	database.DB.Model(&model.GuideSection{}).
		Select("guide_id, COUNT(DISTINCT day_number) as days").
		Where("guide_id IN ?", guideIDs).
		Group("guide_id").
		Find(&dayCounts)
	dayMap := make(map[string]int)
	for _, d := range dayCounts {
		dayMap[d.GuideID] = d.Days
	}

	// 批量查询行程项总数（不包含交通）
	type secCount struct {
		GuideID string
		Count   int64
	}
	var secCounts []secCount
	database.DB.Table("guide_day_items gdi").
		Select("gs.guide_id, COUNT(gdi.id) as count").
		Joins("LEFT JOIN guide_sections gs ON gs.id = gdi.day_id").
		Where("gs.guide_id IN ? AND gdi.section_type != ?", guideIDs, "transport").
		Group("gs.guide_id").
		Find(&secCounts)
	secMap := make(map[string]int64)
	for _, s := range secCounts {
		secMap[s.GuideID] = s.Count
	}

	// 组装返回结果
	items := make([]model.GuideFeedItem, len(guides))
	for i, g := range guides {
		items[i] = model.GuideFeedItem{
			ID:           g.ID,
			UserID:       g.UserID,
			Title:        g.Title,
			CoverImage:   g.CoverImage,
			Destination:  g.Destination,
			IsOriginal:   g.IsOriginal,
			ViewCount:    g.ViewCount,
			LikeCount:    g.LikeCount,
			TripDays:     dayMap[g.ID],
			SectionCount: secMap[g.ID],
			CreatedAt:    g.CreatedAt,
			AuthorName:   userMap[g.UserID].Nickname,
			AuthorAvatar: userMap[g.UserID].AvatarURL,
			IsLiked:      likedSet[g.ID],
		}
	}
	return items, total, nil
}

// CreateGuide 创建攻略（仅基本信息）
func CreateGuide(guide *model.Guide) error {
	return database.DB.Create(guide).Error
}

// CreateGuideWithDays 事务创建攻略 + 每日行程 + 行程项
func CreateGuideWithDays(guide *model.Guide, days []model.GuideSection) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(guide).Error; err != nil {
			return err
		}
		for i := range days {
			days[i].GuideID = guide.ID
			days[i].DayNumber = i + 1
		}
		if len(days) > 0 {
			if err := tx.Create(&days).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetGuideByID 查询攻略详情
func GetGuideByID(id string) (*model.Guide, error) {
	var guide model.Guide
	err := database.DB.First(&guide, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &guide, err
}

// ListGuides 后台攻略列表（所有状态）
func ListGuides(page, pageSize int) ([]model.Guide, int64, error) {
	var guides []model.Guide
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Guide{}).Count(&total)
	err := database.DB.Offset(offset).Limit(pageSize).Find(&guides).Error
	return guides, total, err
}

// UpdateGuideStatus 审核攻略（修改状态）
func UpdateGuideStatus(id string, status int) error {
	return database.DB.Model(&model.Guide{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateGuide 更新攻略基本信息（支持零值更新）
func UpdateGuide(id string, updates map[string]interface{}) error {
	return database.DB.Model(&model.Guide{}).Where("id = ?", id).Updates(updates).Error
}

// IncrementGuideViewCount 增加浏览量
func IncrementGuideViewCount(id string) error {
	return database.DB.Model(&model.Guide{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// GetGuideFavoriteCount 获取攻略收藏数
func GetGuideFavoriteCount(guideID string) int64 {
	var count int64
	database.DB.Model(&model.Favorite{}).Where("target_type = ? AND target_id = ?", "guide", guideID).Count(&count)
	return count
}

// GetGuideCommentCount 获取攻略评论数
func GetGuideCommentCount(guideID string) int64 {
	var count int64
	database.DB.Model(&model.Comment{}).Where("target_type = ? AND target_id = ?", "guide", guideID).Count(&count)
	return count
}

// IsGuideLikedByUser 判断用户是否已点赞攻略
func IsGuideLikedByUser(userID, guideID string) bool {
	var count int64
	database.DB.Model(&model.Favorite{}).Where("user_id = ? AND target_type = ? AND target_id = ?", userID, "guide", guideID).Count(&count)
	return count > 0
}

// ==================== GuideDay（每日行程） ====================

// GetDaysByGuideID 获取攻略的所有每日行程（含行程项）
func GetDaysByGuideID(guideID string) ([]model.GuideSection, error) {
	var days []model.GuideSection
	err := database.DB.Where("guide_id = ?", guideID).
		Order("day_number asc").Find(&days).Error
	if err != nil {
		return nil, err
	}
	// 批量查询每个天的行程项
	dayIDs := make([]string, len(days))
	for i, d := range days {
		dayIDs[i] = d.ID
	}
	if len(dayIDs) > 0 {
		var items []model.GuideDayItem
		database.DB.Where("day_id IN ?", dayIDs).Order("created_at asc").Find(&items)
		itemMap := make(map[string][]model.GuideDayItem)
		for _, it := range items {
			itemMap[it.DayID] = append(itemMap[it.DayID], it)
		}
		for i := range days {
			days[i].Items = itemMap[days[i].ID]
		}
	}
	return days, nil
}

// GetDayByID 获取某天行程
func GetDayByID(id string) (*model.GuideSection, error) {
	var day model.GuideSection
	err := database.DB.First(&day, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &day, nil
}

// GetNextDayNumber 获取攻略的下一个天数序号
func GetNextDayNumber(guideID string) (int, error) {
	var maxDay int
	err := database.DB.Model(&model.GuideSection{}).
		Where("guide_id = ?", guideID).
		Select("COALESCE(MAX(day_number), 0)").
		Scan(&maxDay).Error
	return maxDay + 1, err
}

// CreateDay 创建每日行程（自动分配 DayNumber）
func CreateDay(day *model.GuideSection) error {
	nextNum, err := GetNextDayNumber(day.GuideID)
	if err != nil {
		return err
	}
	day.DayNumber = nextNum
	return database.DB.Create(day).Error
}

// DeleteDay 删除每日行程（级联删除行程项）
func DeleteDay(id string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("day_id = ?", id).Delete(&model.GuideDayItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.GuideSection{}, "id = ?", id).Error
	})
}

// ==================== GuideDayItem（行程项） ====================

// CreateDayItem 创建行程项
func CreateDayItem(item *model.GuideDayItem) error {
	return database.DB.Create(item).Error
}

// UpdateDayItem 更新行程项
func UpdateDayItem(id string, updates map[string]interface{}) error {
	return database.DB.Model(&model.GuideDayItem{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteDayItem 删除行程项
func DeleteDayItem(id string) error {
	return database.DB.Where("id = ?", id).Delete(&model.GuideDayItem{}).Error
}

// GetDayItemByID 获取行程项
func GetDayItemByID(id string) (*model.GuideDayItem, error) {
	var item model.GuideDayItem
	err := database.DB.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ==================== 点赞 / 取消点赞 ====================

// LikeGuide 点赞攻略（幂等：已点赞则直接成功）
func LikeGuide(userID, guideID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 检查是否已点赞
		var count int64
		tx.Model(&model.Favorite{}).
			Where("user_id = ? AND target_type = ? AND target_id = ?", userID, "guide", guideID).
			Count(&count)
		if count > 0 {
			return nil // 已点赞，直接成功
		}
		// 创建点赞记录
		fav := model.Favorite{
			UserID:     userID,
			TargetType: "guide",
			TargetID:   guideID,
		}
		if err := tx.Create(&fav).Error; err != nil {
			return err
		}
		// 点赞数 +1
		return tx.Model(&model.Guide{}).Where("id = ?", guideID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	})
}

// UnlikeGuide 取消点赞攻略（兼容前端传参：支持 guideID 或 Favorite 记录ID）
func UnlikeGuide(userID, guideID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, "guide", guideID).
			Delete(&model.Favorite{})
		if result.RowsAffected == 0 {
			// 未匹配到，尝试按记录ID删除
			result = tx.Where("id = ? AND user_id = ?", guideID, userID).
				Delete(&model.Favorite{})
			if result.RowsAffected == 0 {
				return nil // 未点赞也视为成功（幂等）
			}
		}
		// 点赞数 -1（确保不小于 0）
		return tx.Model(&model.Guide{}).Where("id = ?", guideID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
	})
}

// GetItemsByDayID 获取某天的所有行程项
func GetItemsByDayID(dayID string) ([]model.GuideDayItem, error) {
	var items []model.GuideDayItem
	err := database.DB.Where("day_id = ?", dayID).
		Order("created_at asc").Find(&items).Error
	return items, err
}

package repository

import (
	"sort"

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
	// 批量获取当前用户的点赞状态（action=1 点赞，排除收藏）
	likedSet := make(map[string]bool)
	if userID != "" {
		var favs []model.Favorite
		database.DB.Where("user_id = ? AND target_type = ? AND target_id IN ? AND action = ?", userID, "guide", guideIDs, 1).Find(&favs)
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
			Status:       g.Status,
			CreatedAt:    g.CreatedAt,
			AuthorName:   userMap[g.UserID].Nickname,
			AuthorAvatar: userMap[g.UserID].AvatarURL,
			IsLiked:      likedSet[g.ID],
		}
	}
	return items, total, nil
}

// GetPublicFeed 获取公开内容瀑布流（已发布攻略 + 公开行程，合并按时间/热度倒序，支持目的地/tab筛选）
// tab 取值：recommend(推荐/默认) hot(热门) latest(最新) domestic(国内) overseas(国外)
func GetPublicFeed(page, pageSize int, destination, keyword, userID, tab string) ([]model.FeedItem, int64, error) {
	// 1. 查询已发布的攻略
	guideQuery := database.DB.Model(&model.Guide{}).Where("status = ?", 1)
	// 2. 查询公开的行程
	tripQuery := database.DB.Model(&model.Trip{}).Where("is_public = ?", 1)

	if destination != "" {
		guideQuery = guideQuery.Where("destination LIKE ?", "%"+destination+"%")
		tripQuery = tripQuery.Where("destinations LIKE ?", "%"+destination+"%")
	}

	// 关键词搜索：匹配标题/目的地/简介
	if keyword != "" {
		kw := "%" + keyword + "%"
		guideQuery = guideQuery.Where("title LIKE ? OR destination LIKE ? OR summary LIKE ?", kw, kw, kw)
		tripQuery = tripQuery.Where("title LIKE ? OR destinations LIKE ? OR summary LIKE ?", kw, kw, kw)
	}

	// 按 tab 筛选及排序
	var guideOrder, tripOrder string
	switch tab {
	case "hot":
		guideOrder = "like_count desc"
		tripOrder = "like_count desc"
	case "domestic":
		guideQuery = guideQuery.Where("is_overseas = ?", 0)
		tripQuery = tripQuery.Where("is_overseas = ?", 0)
		guideOrder = "created_at desc"
		tripOrder = "created_at desc"
	case "overseas":
		guideQuery = guideQuery.Where("is_overseas = ?", 1)
		tripQuery = tripQuery.Where("is_overseas = ?", 1)
		guideOrder = "created_at desc"
		tripOrder = "created_at desc"
	default: // recommend / latest
		guideOrder = "created_at desc"
		tripOrder = "created_at desc"
	}

	var guideTotal, tripTotal int64
	guideQuery.Count(&guideTotal)
	tripQuery.Count(&tripTotal)

	var guides []model.Guide
	var trips []model.Trip
	guideQuery.Order(guideOrder).Find(&guides)
	tripQuery.Order(tripOrder).Find(&trips)

	// 批量查询攻略的天数和行程项统计
	guideIDs := make([]string, len(guides))
	for i, g := range guides {
		guideIDs[i] = g.ID
	}

	type dayCount struct {
		GuideID string
		Days    int
	}
	var dayCounts []dayCount
	if len(guideIDs) > 0 {
		database.DB.Model(&model.GuideSection{}).
			Select("guide_id, COUNT(DISTINCT day_number) as days").
			Where("guide_id IN ?", guideIDs).
			Group("guide_id").
			Find(&dayCounts)
	}
	dayMap := make(map[string]int)
	for _, d := range dayCounts {
		dayMap[d.GuideID] = d.Days
	}

	type secCount struct {
		GuideID string
		Count   int64
	}
	var secCounts []secCount
	if len(guideIDs) > 0 {
		database.DB.Table("guide_day_items gdi").
			Select("gs.guide_id, COUNT(gdi.id) as count").
			Joins("LEFT JOIN guide_sections gs ON gs.id = gdi.day_id").
			Where("gs.guide_id IN ? AND gdi.section_type != ?", guideIDs, "transport").
			Group("gs.guide_id").
			Find(&secCounts)
	}
	secMap := make(map[string]int64)
	for _, s := range secCounts {
		secMap[s.GuideID] = s.Count
	}

	// 3. 合并为 FeedItem 切片
	allItems := make([]model.FeedItem, 0, len(guides)+len(trips))
	for _, g := range guides {
		allItems = append(allItems, model.FeedItem{
			ID:           g.ID,
			UserID:       g.UserID,
			Title:        g.Title,
			CoverImage:   g.CoverImage,
			Destinations: []string{g.Destination},
			Summary:      g.Summary,
			ItemType:     "guide",
			IsOverseas:   g.IsOverseas,
			ViewCount:    g.ViewCount,
			LikeCount:    g.LikeCount,
			TripDays:     dayMap[g.ID],
			SectionCount: secMap[g.ID],
			Status:       g.Status,
			CreatedAt:    g.CreatedAt,
		})
	}
	// 批量查询行程的天数和行程项统计
	tripIDs := make([]string, len(trips))
	for i, t := range trips {
		tripIDs[i] = t.ID
	}

	type tripDayCount struct {
		TripID string
		Days   int
	}
	var tripDayCounts []tripDayCount
	if len(tripIDs) > 0 {
		database.DB.Model(&model.TripDay{}).
			Select("trip_id, COUNT(DISTINCT day_number) as days").
			Where("trip_id IN ?", tripIDs).
			Group("trip_id").
			Find(&tripDayCounts)
	}
	tripDayMap := make(map[string]int)
	for _, d := range tripDayCounts {
		tripDayMap[d.TripID] = d.Days
	}

	type tripSecCount struct {
		TripID string
		Count  int64
	}
	var tripSecCounts []tripSecCount
	if len(tripIDs) > 0 {
		database.DB.Table("trip_items ti").
			Select("td.trip_id, COUNT(ti.id) as count").
			Joins("LEFT JOIN trip_days td ON td.id = ti.trip_day_id").
			Where("td.trip_id IN ?", tripIDs).
			Group("td.trip_id").
			Find(&tripSecCounts)
	}
	tripSecMap := make(map[string]int64)
	for _, s := range tripSecCounts {
		tripSecMap[s.TripID] = s.Count
	}

	for _, t := range trips {
		allItems = append(allItems, model.FeedItem{
			ID:           t.ID,
			UserID:       t.UserID,
			Title:        t.Title,
			CoverImage:   t.CoverImage,
			Destinations: t.Destinations,
			Summary:      t.Summary,
			ItemType:     "trip",
			IsOverseas:   t.IsOverseas,
			ViewCount:    t.ViewCount,
			LikeCount:    t.LikeCount,
			TripDays:     tripDayMap[t.ID],
			SectionCount: tripSecMap[t.ID],
			Status:       t.Status,
			IsPublic:     t.IsPublic,
			CreatedAt:    t.CreatedAt,
		})
	}

	// 4. 合并排序（热门按点赞数倒序，其余按创建时间倒序）
	if tab == "hot" {
		sort.Slice(allItems, func(i, j int) bool {
			return allItems[i].LikeCount > allItems[j].LikeCount
		})
	} else {
		sort.Slice(allItems, func(i, j int) bool {
			return allItems[i].CreatedAt.After(allItems[j].CreatedAt)
		})
	}

	total := guideTotal + tripTotal
	offset := (page - 1) * pageSize
	if offset > len(allItems) {
		return []model.FeedItem{}, total, nil
	}
	end := offset + pageSize
	if end > len(allItems) {
		end = len(allItems)
	}
	items := allItems[offset:end]

	// 5. 批量获取作者信息
	userIDs := make([]string, 0)
	itemIDs := make([]string, 0)
	userIDSet := make(map[string]bool)
	for i := range items {
		if !userIDSet[items[i].UserID] {
			userIDs = append(userIDs, items[i].UserID)
			userIDSet[items[i].UserID] = true
		}
		itemIDs = append(itemIDs, items[i].ID)
	}

	var users []model.User
	database.DB.Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[string]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// 6. 批量获取当前用户的点赞状态（攻略和行程共用 Favorite 表，action=1 点赞，排除收藏）
	likedSet := make(map[string]bool)
	if userID != "" {
		var favs []model.Favorite
		database.DB.Where("user_id = ? AND target_id IN ? AND action = ?", userID, itemIDs, 1).Find(&favs)
		for _, f := range favs {
			likedSet[f.TargetID] = true
		}
	}

	for i := range items {
		items[i].AuthorName = userMap[items[i].UserID].Nickname
		items[i].AuthorAvatar = userMap[items[i].UserID].AvatarURL
		items[i].IsLiked = likedSet[items[i].ID]
	}

	return items, total, nil
}

// GetUserFeed 获取用户发布的公开内容（已发布攻略+公开行程，合并按时间倒序）
func GetUserFeed(userID string, page, pageSize int) ([]model.FeedItem, int64, error) {
	// 1. 查询已发布的攻略
	var guides []model.Guide
	var guideTotal int64
	guideQuery := database.DB.Model(&model.Guide{}).Where("user_id = ? AND status = 1", userID)
	guideQuery.Count(&guideTotal)
	guideQuery.Order("created_at desc").Find(&guides)

	// 2. 查询公开的行程
	var trips []model.Trip
	var tripTotal int64
	tripQuery := database.DB.Model(&model.Trip{}).Where("user_id = ? AND is_public = 1", userID)
	tripQuery.Count(&tripTotal)
	tripQuery.Order("created_at desc").Find(&trips)

	// 3. 查询已发布的公开搭子
	var partners []model.Partner
	var partnerTotal int64
	partnerQuery := database.DB.Model(&model.Partner{}).Where("user_id = ? AND is_draft = 0 AND is_public = 1", userID)
	partnerQuery.Count(&partnerTotal)
	partnerQuery.Order("created_at desc").Preload("Days").Find(&partners)

	return mergeUserFeedItems(guides, trips, partners, guideTotal+tripTotal+partnerTotal, page, pageSize)
}

// ListMyNotes 我的全部笔记（攻略+行程+搭子，含草稿/私密，合并按时间倒序），供"我的笔记-全部"Tab使用
func ListMyNotes(userID string, page, pageSize int) ([]model.FeedItem, int64, error) {
	// 1. 查询我的攻略（全部状态）
	var guides []model.Guide
	var guideTotal int64
	guideQuery := database.DB.Model(&model.Guide{}).Where("user_id = ?", userID)
	guideQuery.Count(&guideTotal)
	guideQuery.Order("created_at desc").Find(&guides)

	// 2. 查询我的行程（全部状态）
	var trips []model.Trip
	var tripTotal int64
	tripQuery := database.DB.Model(&model.Trip{}).Where("user_id = ?", userID)
	tripQuery.Count(&tripTotal)
	tripQuery.Order("created_at desc").Find(&trips)

	// 3. 查询我的搭子（全部状态）
	var partners []model.Partner
	var partnerTotal int64
	partnerQuery := database.DB.Model(&model.Partner{}).Where("user_id = ?", userID)
	partnerQuery.Count(&partnerTotal)
	partnerQuery.Order("created_at desc").Preload("Days").Find(&partners)

	return mergeUserFeedItems(guides, trips, partners, guideTotal+tripTotal+partnerTotal, page, pageSize)
}

// mergeUserFeedItems 合并用户的攻略+行程+搭子：补全天数/行程项统计，按时间倒序后手动分页
func mergeUserFeedItems(guides []model.Guide, trips []model.Trip, partners []model.Partner, total int64, page, pageSize int) ([]model.FeedItem, int64, error) {
	// 批量查询攻略的天数和行程项统计
	guideIDs := make([]string, len(guides))
	for i, g := range guides {
		guideIDs[i] = g.ID
	}
	type dayCount struct {
		GuideID string
		Days    int
	}
	var dayCounts []dayCount
	if len(guideIDs) > 0 {
		database.DB.Model(&model.GuideSection{}).
			Select("guide_id, COUNT(DISTINCT day_number) as days").
			Where("guide_id IN ?", guideIDs).
			Group("guide_id").
			Find(&dayCounts)
	}
	dayMap := make(map[string]int)
	for _, d := range dayCounts {
		dayMap[d.GuideID] = d.Days
	}
	type secCount struct {
		GuideID string
		Count   int64
	}
	var secCounts []secCount
	if len(guideIDs) > 0 {
		database.DB.Table("guide_day_items gdi").
			Select("gs.guide_id, COUNT(gdi.id) as count").
			Joins("LEFT JOIN guide_sections gs ON gs.id = gdi.day_id").
			Where("gs.guide_id IN ? AND gdi.section_type != ?", guideIDs, "transport").
			Group("gs.guide_id").
			Find(&secCounts)
	}
	secMap := make(map[string]int64)
	for _, s := range secCounts {
		secMap[s.GuideID] = s.Count
	}

	// 批量查询行程的天数和行程项统计
	tripIDs := make([]string, len(trips))
	for i, t := range trips {
		tripIDs[i] = t.ID
	}
	type tripDayCount struct {
		TripID string
		Days   int
	}
	var tripDayCounts []tripDayCount
	if len(tripIDs) > 0 {
		database.DB.Model(&model.TripDay{}).
			Select("trip_id, COUNT(DISTINCT day_number) as days").
			Where("trip_id IN ?", tripIDs).
			Group("trip_id").
			Find(&tripDayCounts)
	}
	tripDayMap := make(map[string]int)
	for _, d := range tripDayCounts {
		tripDayMap[d.TripID] = d.Days
	}
	type tripSecCount struct {
		TripID string
		Count  int64
	}
	var tripSecCounts []tripSecCount
	if len(tripIDs) > 0 {
		database.DB.Table("trip_items ti").
			Select("td.trip_id, COUNT(ti.id) as count").
			Joins("LEFT JOIN trip_days td ON td.id = ti.trip_day_id").
			Where("td.trip_id IN ?", tripIDs).
			Group("td.trip_id").
			Find(&tripSecCounts)
	}
	tripSecMap := make(map[string]int64)
	for _, s := range tripSecCounts {
		tripSecMap[s.TripID] = s.Count
	}

	// 4. 合并为 FeedItem 切片（攻略+行程+搭子）
	allItems := make([]model.FeedItem, 0, len(guides)+len(trips)+len(partners))
	for _, g := range guides {
		allItems = append(allItems, model.FeedItem{
			ID:           g.ID,
			UserID:       g.UserID,
			Title:        g.Title,
			CoverImage:   g.CoverImage,
			Destinations: []string{g.Destination},
			Summary:      g.Summary,
			ItemType:     "guide",
			IsOverseas:   g.IsOverseas,
			ViewCount:    g.ViewCount,
			LikeCount:    g.LikeCount,
			TripDays:     dayMap[g.ID],
			SectionCount: secMap[g.ID],
			Status:       g.Status,
			CreatedAt:    g.CreatedAt,
		})
	}
	for _, t := range trips {
		allItems = append(allItems, model.FeedItem{
			ID:           t.ID,
			UserID:       t.UserID,
			Title:        t.Title,
			CoverImage:   t.CoverImage,
			Destinations: t.Destinations,
			Summary:      t.Summary,
			ItemType:     "trip",
			IsOverseas:   t.IsOverseas,
			ViewCount:    t.ViewCount,
			LikeCount:    t.LikeCount,
			TripDays:     tripDayMap[t.ID],
			SectionCount: tripSecMap[t.ID],
			Status:       t.Status,
			IsPublic:     t.IsPublic,
			CreatedAt:    t.CreatedAt,
		})
	}
	for _, p := range partners {
		allItems = append(allItems, model.FeedItem{
			ID:           p.ID,
			UserID:       p.UserID,
			Title:        p.Title,
			CoverImage:   p.Cover,
			Destinations: []string{p.Destination},
			Summary:      p.Desc,
			ItemType:     "partner",
			ViewCount:    p.ViewCount,
			LikeCount:    p.LikeCount,
			TripDays:     len(p.Days),
			Status:       p.Status,
			IsDraft:      p.IsDraft,
			IsPublic:     p.IsPublic,
			CreatedAt:    p.CreatedAt,
		})
	}

	// 5. 按时间倒序排序
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].CreatedAt.After(allItems[j].CreatedAt)
	})

	// 6. 手动分页
	offset := (page - 1) * pageSize
	if offset > len(allItems) {
		return []model.FeedItem{}, total, nil
	}
	end := offset + pageSize
	if end > len(allItems) {
		end = len(allItems)
	}
	return allItems[offset:end], total, nil
}

// ListMyGuides 我的攻略列表（含天数、行程项统计；status>=0 时按状态筛选，-1 返回全部）
func ListMyGuides(userID string, page, pageSize, status int) ([]model.GuideFeedItem, int64, error) {
	var guides []model.Guide
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Guide{}).Where("user_id = ?", userID)
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	listQuery := database.DB.Where("user_id = ?", userID)
	if status >= 0 {
		listQuery = listQuery.Where("status = ?", status)
	}
	err := listQuery.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&guides).Error
	if err != nil {
		return nil, 0, err
	}

	// 批量查询天数统计
	guideIDs := make([]string, len(guides))
	for i, g := range guides {
		guideIDs[i] = g.ID
	}

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
			Status:       g.Status,
			CreatedAt:    g.CreatedAt,
		}
	}
	return items, total, nil
}

// DeleteGuideCascade 删除攻略（级联删除每日行程及行程项）
func DeleteGuideCascade(id string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var dayIDs []string
		if err := tx.Model(&model.GuideSection{}).Where("guide_id = ?", id).Pluck("id", &dayIDs).Error; err != nil {
			return err
		}
		if len(dayIDs) > 0 {
			if err := tx.Where("day_id IN ?", dayIDs).Delete(&model.GuideDayItem{}).Error; err != nil {
				return err
			}
			if err := tx.Where("guide_id = ?", id).Delete(&model.GuideSection{}).Error; err != nil {
				return err
			}
		}
		return tx.Where("id = ?", id).Delete(&model.Guide{}).Error
	})
}

// ListUserPublishedGuides 他人已发布的攻略列表
func ListUserPublishedGuides(userID string, page, pageSize int) ([]model.GuideFeedItem, int64, error) {
	var guides []model.Guide
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Guide{}).Where("user_id = ? AND status = 1", userID).Count(&total)
	err := database.DB.Where("user_id = ? AND status = 1", userID).
		Order("created_at desc").Offset(offset).Limit(pageSize).Find(&guides).Error
	if err != nil {
		return nil, 0, err
	}

	// 批量查询天数统计
	guideIDs := make([]string, len(guides))
	for i, g := range guides {
		guideIDs[i] = g.ID
	}

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
		Where("gs.guide_id IN ?", guideIDs).
		Group("gs.guide_id").
		Find(&secCounts)
	secMap := make(map[string]int64)
	for _, s := range secCounts {
		secMap[s.GuideID] = s.Count
	}

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
			Status:       g.Status,
			CreatedAt:    g.CreatedAt,
		}
	}
	return items, total, nil
}

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

// ListGuides 后台攻略列表（所有状态，支持状态筛选与目的地/标题模糊搜索）
func ListGuides(page, pageSize int, status int, destination string, title string) ([]model.Guide, int64, error) {
	var guides []model.Guide
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Guide{})
	// status 仅筛选有效值（-1 或未传表示全部）
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	// 目的地关键词模糊搜索
	if destination != "" {
		query = query.Where("destination LIKE ?", "%"+destination+"%")
	}
	// 标题关键词模糊搜索
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&guides).Error
	return guides, total, err
}

// UpdateGuideStatus 审核攻略（修改状态）
func UpdateGuideStatus(id string, status int) error {
	return database.DB.Model(&model.Guide{}).Where("id = ?", id).Update("status", status).Error
}

// ListAllGuides 全量攻略列表（用于存量数据回填）
func ListAllGuides() ([]model.Guide, error) {
	var guides []model.Guide
	err := database.DB.Model(&model.Guide{}).Find(&guides).Error
	return guides, err
}

// UpdateGuide 更新攻略基本信息（支持零值更新）
func UpdateGuide(id string, updates map[string]interface{}) error {
	return database.DB.Model(&model.Guide{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateGuideWithDays 事务更新攻略基本信息并全量替换每日行程（删除旧行程后重建）
func UpdateGuideWithDays(id string, updates map[string]interface{}, days []model.GuideSection) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&model.Guide{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
		}
		// 删除旧行程日及行程项
		var dayIDs []string
		if err := tx.Model(&model.GuideSection{}).Where("guide_id = ?", id).Pluck("id", &dayIDs).Error; err != nil {
			return err
		}
		if len(dayIDs) > 0 {
			if err := tx.Where("day_id IN ?", dayIDs).Delete(&model.GuideDayItem{}).Error; err != nil {
				return err
			}
			if err := tx.Where("guide_id = ?", id).Delete(&model.GuideSection{}).Error; err != nil {
				return err
			}
		}
		// 重建行程日
		for i := range days {
			days[i].ID = ""
			days[i].GuideID = id
			days[i].DayNumber = i + 1
			for j := range days[i].Items {
				days[i].Items[j].ID = ""
				days[i].Items[j].DayID = ""
			}
		}
		if len(days) > 0 {
			if err := tx.Create(&days).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// IncrementGuideViewCount 增加浏览量
func IncrementGuideViewCount(id string) error {
	return database.DB.Model(&model.Guide{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// GetGuideFavoriteCount 获取攻略收藏数
func GetGuideFavoriteCount(guideID string) int64 {
	var count int64
	database.DB.Model(&model.Favorite{}).Where("target_type = ? AND target_id = ? AND action = 2", "guide", guideID).Count(&count)
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
	database.DB.Model(&model.Favorite{}).Where("user_id = ? AND target_type = ? AND target_id = ? AND action = 1", userID, "guide", guideID).Count(&count)
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

// LikeGuide 点赞攻略（幂等：已点赞则直接成功；点赞记录 action=1，与收藏分离）
func LikeGuide(userID, guideID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 检查是否已点赞
		var count int64
		tx.Model(&model.Favorite{}).
			Where("user_id = ? AND target_type = ? AND target_id = ? AND action = 1", userID, "guide", guideID).
			Count(&count)
		if count > 0 {
			return nil // 已点赞，直接成功
		}
		// 创建点赞记录
		fav := model.Favorite{
			UserID:     userID,
			TargetType: "guide",
			TargetID:   guideID,
			Action:     1,
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
		result := tx.Where("user_id = ? AND target_type = ? AND target_id = ? AND action = 1", userID, "guide", guideID).
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

package repository

import (
	"time"

	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// ==================== Trip ====================

// CreateTrip 创建行程
func CreateTrip(trip *model.Trip) error {
	return database.DB.Create(trip).Error
}

// CountTodayAITrips 统计用户今日AI生成的行程数
func CountTodayAITrips(userID string) (int64, error) {
	var count int64
	start := time.Now().Truncate(24 * time.Hour)
	err := database.DB.Model(&model.Trip{}).
		Where("user_id = ? AND is_ai = 1 AND created_at >= ?", userID, start).
		Count(&count).Error
	return count, err
}

// GetTripByID 获取行程详情（含行程日+行程项+同行者）
func GetTripByID(id string) (*model.Trip, error) {
	var trip model.Trip
	err := database.DB.
		Preload("Days.Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc")
		}).
		Preload("Members").
		First(&trip, "id = ?", id).Error
	return &trip, err
}

// UpdateTrip 更新行程基本信息
func UpdateTrip(id string, updates map[string]interface{}) error {
	return database.DB.Model(&model.Trip{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateTripWithDays 事务更新行程基本信息并全量替换行程日（删除旧行程日后重建）
func UpdateTripWithDays(id string, updates map[string]interface{}, days []model.TripDay) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&model.Trip{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
		}
		// 删除旧行程日及行程项
		var dayIDs []string
		if err := tx.Model(&model.TripDay{}).Where("trip_id = ?", id).Pluck("id", &dayIDs).Error; err != nil {
			return err
		}
		if len(dayIDs) > 0 {
			if err := tx.Where("trip_day_id IN ?", dayIDs).Delete(&model.TripItem{}).Error; err != nil {
				return err
			}
			if err := tx.Where("trip_id = ?", id).Delete(&model.TripDay{}).Error; err != nil {
				return err
			}
		}
		// 重建行程日
		for i := range days {
			days[i].ID = ""
			days[i].TripID = id
			for j := range days[i].Items {
				days[i].Items[j].ID = ""
				days[i].Items[j].TripDayID = ""
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

// DeleteTrip 软删除行程
func DeleteTrip(id string) error {
	return database.DB.Where("id = ?", id).Delete(&model.Trip{}).Error
}

// DeleteTripCascade 删除行程（级联删除行程日、行程项及同行者）
func DeleteTripCascade(id string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var dayIDs []string
		if err := tx.Model(&model.TripDay{}).Where("trip_id = ?", id).Pluck("id", &dayIDs).Error; err != nil {
			return err
		}
		if len(dayIDs) > 0 {
			if err := tx.Where("trip_day_id IN ?", dayIDs).Delete(&model.TripItem{}).Error; err != nil {
				return err
			}
			if err := tx.Where("trip_id = ?", id).Delete(&model.TripDay{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("trip_id = ?", id).Delete(&model.TripMember{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.Trip{}).Error
	})
}

// GetTripFavoriteCount 获取行程收藏数
func GetTripFavoriteCount(tripID string) int64 {
	var count int64
	database.DB.Model(&model.Favorite{}).Where("target_type = ? AND target_id = ? AND action = 2", "trip", tripID).Count(&count)
	return count
}

// GetTripCommentCount 获取行程评论数
func GetTripCommentCount(tripID string) int64 {
	var count int64
	database.DB.Model(&model.Comment{}).Where("target_type = ? AND target_id = ?", "trip", tripID).Count(&count)
	return count
}

// ListUserTrips 我的行程列表（status>=0 时按状态筛选，-1 返回全部）
func ListUserTrips(userID string, page, pageSize, status int) ([]model.Trip, int64, error) {
	var trips []model.Trip
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Trip{}).Where("user_id = ?", userID)
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	listQuery := database.DB.Where("user_id = ?", userID)
	if status >= 0 {
		listQuery = listQuery.Where("status = ?", status)
	}
	err := listQuery.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&trips).Error
	return trips, total, err
}

// ListUserPublishedTrips 他人已公开的行程列表
func ListUserPublishedTrips(userID string, page, pageSize int) ([]model.Trip, int64, error) {
	var trips []model.Trip
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Trip{}).Where("user_id = ? AND is_public = 1", userID).Count(&total)
	err := database.DB.Where("user_id = ? AND is_public = 1", userID).
		Order("created_at desc").Offset(offset).Limit(pageSize).Find(&trips).Error
	return trips, total, err
}

// ListPublicTrips 公开行程列表
func ListPublicTrips(page, pageSize int, destination string) ([]model.Trip, int64, error) {
	var trips []model.Trip
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Trip{}).Where("is_public = ?", 1)
	if destination != "" {
		query = query.Where("destinations LIKE ?", "%"+destination+"%")
	}
	query.Count(&total)
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&trips).Error
	return trips, total, err
}

// ==================== TripDay ====================

// CreateTripDay 创建行程日
func CreateTripDay(day *model.TripDay) error {
	return database.DB.Create(day).Error
}

// UpdateTripDay 更新行程日
func UpdateTripDay(id string, updates map[string]interface{}) error {
	return database.DB.Model(&model.TripDay{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteTripDay 删除行程日（级联删除下属行程项）
func DeleteTripDay(id string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("trip_day_id = ?", id).Delete(&model.TripItem{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.TripDay{}).Error
	})
}

// ==================== TripItem ====================

// CreateTripItem 创建行程项
func CreateTripItem(item *model.TripItem) error {
	return database.DB.Create(item).Error
}

// UpdateTripItem 更新行程项
func UpdateTripItem(id string, updates map[string]interface{}) error {
	return database.DB.Model(&model.TripItem{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteTripItem 删除行程项
func DeleteTripItem(id string) error {
	return database.DB.Where("id = ?", id).Delete(&model.TripItem{}).Error
}

// BatchCreateTripItems 批量创建行程项
func BatchCreateTripItems(items []model.TripItem) error {
	return database.DB.Create(&items).Error
}

// ReorderTripItems 重新排序行程项
func ReorderTripItems(tripDayID string, itemIDs []string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for i, id := range itemIDs {
			if err := tx.Model(&model.TripItem{}).Where("id = ? AND trip_day_id = ?", id, tripDayID).
				Update("sort_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== TripMember ====================

// AddTripMember 添加同行者
func AddTripMember(member *model.TripMember) error {
	return database.DB.Create(member).Error
}

// RemoveTripMember 移除同行者
func RemoveTripMember(id string) error {
	return database.DB.Where("id = ?", id).Delete(&model.TripMember{}).Error
}

// UpdateTripMemberRole 更新同行者角色
func UpdateTripMemberRole(id string, role string) error {
	return database.DB.Model(&model.TripMember{}).Where("id = ?", id).Update("role", role).Error
}

// GetTripItemCounts 批量查询多个行程的行程项总数（按行程ID分组）
func GetTripItemCounts(tripIDs []string) map[string]int64 {
	if len(tripIDs) == 0 {
		return nil
	}
	type result struct {
		TripID string
		Count  int64
	}
	var rows []result
	database.DB.Table("trip_items ti").
		Select("td.trip_id, COUNT(ti.id) as count").
		Joins("JOIN trip_days td ON ti.trip_day_id = td.id").
		Where("td.trip_id IN ?", tripIDs).
		Group("td.trip_id").
		Find(&rows)
	m := make(map[string]int64, len(rows))
	for _, r := range rows {
		m[r.TripID] = r.Count
	}
	return m
}

// UpdateTripStatus 审核行程（修改状态）
func UpdateTripStatus(id string, status int) error {
	return database.DB.Model(&model.Trip{}).Where("id = ?", id).Update("status", status).Error
}

// ListAllTrips 后台行程列表（全部用户，支持状态/关键词筛选）
func ListAllTrips(page, pageSize, status int, keyword string) ([]model.Trip, int64, error) {
	var trips []model.Trip
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Trip{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR destinations LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&trips).Error
	return trips, total, err
}

// IncrementTripViewCount 增加行程浏览量
func IncrementTripViewCount(id string) error {
	return database.DB.Model(&model.Trip{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// ==================== 点赞 / 取消点赞 ====================

// IsTripLikedByUser 判断用户是否已点赞行程
func IsTripLikedByUser(userID, tripID string) bool {
	var count int64
	database.DB.Model(&model.Favorite{}).
		Where("user_id = ? AND target_type = ? AND target_id = ? AND action = 1", userID, "trip", tripID).
		Count(&count)
	return count > 0
}

// LikeTrip 点赞行程（幂等：已点赞则直接成功；点赞记录 action=1，与收藏分离）
func LikeTrip(userID, tripID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 检查是否已点赞
		var count int64
		tx.Model(&model.Favorite{}).
			Where("user_id = ? AND target_type = ? AND target_id = ? AND action = 1", userID, "trip", tripID).
			Count(&count)
		if count > 0 {
			return nil // 已点赞，直接成功
		}
		// 创建点赞记录
		fav := model.Favorite{
			UserID:     userID,
			TargetType: "trip",
			TargetID:   tripID,
			Action:     1,
		}
		if err := tx.Create(&fav).Error; err != nil {
			return err
		}
		// 点赞数 +1
		return tx.Model(&model.Trip{}).Where("id = ?", tripID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	})
}

// UnlikeTrip 取消点赞行程（幂等：未点赞也视为成功）
func UnlikeTrip(userID, tripID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND target_type = ? AND target_id = ? AND action = 1", userID, "trip", tripID).
			Delete(&model.Favorite{})
		if result.RowsAffected == 0 {
			return nil // 未点赞也视为成功（幂等）
		}
		// 点赞数 -1（确保不小于 0）
		return tx.Model(&model.Trip{}).Where("id = ?", tripID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
	})
}

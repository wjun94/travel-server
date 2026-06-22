package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// CreateTrip 创建行程
func CreateTrip(trip *model.Trip) error {
	return database.DB.Create(trip).Error
}

// GetTripByID 获取行程详情（含协作者）
func GetTripByID(id uint) (*model.Trip, error) {
	var trip model.Trip
	err := database.DB.Preload("Collaborators").First(&trip, id).Error
	return &trip, err
}

// UpdateTrip 更新行程（含乐观锁版本号）
func UpdateTrip(trip *model.Trip) error {
	// 使用 version 字段避免并发覆盖
	result := database.DB.Model(&model.Trip{}).
		Where("id = ? AND version = ?", trip.ID, trip.Version).
		Updates(map[string]interface{}{
			"daily_plans": trip.DailyPlans,
			"version":     trip.Version + 1,
		})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 版本冲突
	}
	return result.Error
}

// AddCollaborator 添加协作者
func AddCollaborator(tripID, userID uint, perm int) error {
	c := model.TripCollaborator{TripID: tripID, UserID: userID, Permission: perm}
	return database.DB.Create(&c).Error
}

// RemoveCollaborator 移除协作者
func RemoveCollaborator(tripID, userID uint) error {
	return database.DB.Where("trip_id = ? AND user_id = ?", tripID, userID).Delete(&model.TripCollaborator{}).Error
}

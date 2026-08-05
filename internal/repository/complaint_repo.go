package repository

import (
	"time"

	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// CreateComplaint 提交投诉
func CreateComplaint(c *model.Complaint) error {
	return database.DB.Create(c).Error
}

// ListComplaints 投诉列表（分页，支持状态/目标类型筛选）
func ListComplaints(page, pageSize int, status *int, targetType string) ([]model.Complaint, int64, error) {
	var complaints []model.Complaint
	var total int64
	query := database.DB.Model(&model.Complaint{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&complaints).Error
	return complaints, total, err
}

// ListUserComplaints 我的投诉列表（分页，仅本人）
func ListUserComplaints(userID string, page, pageSize int) ([]model.Complaint, int64, error) {
	var complaints []model.Complaint
	var total int64
	query := database.DB.Model(&model.Complaint{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&complaints).Error
	return complaints, total, err
}

// GetUserComplaint 我的投诉详情（仅本人）
func GetUserComplaint(id, userID string) (*model.Complaint, error) {
	var c model.Complaint
	err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&c).Error
	return &c, err
}

// GetComplaintByID 获取投诉详情
func GetComplaintByID(id string) (*model.Complaint, error) {
	var c model.Complaint
	err := database.DB.First(&c, "id = ?", id).Error
	return &c, err
}

// UpdateComplaintStatus 处理投诉（设置状态、处理备注与回复）
func UpdateComplaintStatus(id string, status int, handleNote, reply string) error {
	return database.DB.Model(&model.Complaint{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"handle_note": handleNote,
			"reply":       reply,
			"handled_at":  time.Now(),
		}).Error
}

// DeleteComplaint 删除投诉
func DeleteComplaint(id string) error {
	return database.DB.Where("id = ?", id).Delete(&model.Complaint{}).Error
}

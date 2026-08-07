package repository

import (
	"time"

	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// CreateSysMessage 创建系统消息记录
func CreateSysMessage(msg *model.SysMessage) error {
	return database.DB.Create(msg).Error
}

// ListSysMessages 系统消息列表（分页，支持状态筛选）
func ListSysMessages(page, pageSize int, status *int) ([]model.SysMessage, int64, error) {
	var list []model.SysMessage
	var total int64
	query := database.DB.Model(&model.SysMessage{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

// GetSysMessageByID 获取系统消息详情
func GetSysMessageByID(id string) (*model.SysMessage, error) {
	var msg model.SysMessage
	err := database.DB.First(&msg, "id = ?", id).Error
	return &msg, err
}

// ListPendingSysMessages 获取所有到期的待发送消息（定时发送扫描用）
func ListPendingSysMessages() ([]model.SysMessage, error) {
	var list []model.SysMessage
	err := database.DB.Where("status = 0 AND send_time <= ?", time.Now()).
		Order("send_time ASC").Find(&list).Error
	return list, err
}

// MarkSysMessageSent 标记消息已发送并记录实际送达人数
func MarkSysMessageSent(id string, sentCount int) error {
	return database.DB.Model(&model.SysMessage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     1,
			"sent_at":    time.Now(),
			"sent_count": sentCount,
		}).Error
}

// CancelSysMessage 取消待发送的消息（仅 status=0 可取消）
func CancelSysMessage(id string) (bool, error) {
	res := database.DB.Model(&model.SysMessage{}).
		Where("id = ? AND status = 0", id).
		Update("status", 2)
	return res.RowsAffected > 0, res.Error
}

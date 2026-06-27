package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// GetChecklistsByUser 获取用户的备忘清单
func GetChecklistsByUser(userID string) ([]model.Checklist, error) {
	var lists []model.Checklist
	err := database.DB.Where("user_id = ?", userID).Preload("Items").Find(&lists).Error
	return lists, err
}

// CreateChecklist 创建清单
func CreateChecklist(cl *model.Checklist) error {
	return database.DB.Create(cl).Error
}

// UpdateChecklistItem 更新清单条目勾选状态
func UpdateChecklistItem(id string, checked int) error {
	return database.DB.Model(&model.ChecklistItem{}).Where("id = ?", id).Update("checked", checked).Error
}

package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
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

// GetChecklistCategories 获取系统预置的备忘清单分类（含条目）
func GetChecklistCategories() ([]model.ChecklistCategory, error) {
	var cats []model.ChecklistCategory
	err := database.DB.Order("sort_order asc").Preload("Items").Find(&cats).Error
	return cats, err
}

// DeleteChecklist 删除备忘清单（级联删除条目）
func DeleteChecklist(id, userID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("checklist_id = ?", id).Delete(&model.ChecklistItem{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Checklist{}).Error
	})
}

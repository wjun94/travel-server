package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// GetChecklistsByUser 获取用户的备忘清单（分页）
func GetChecklistsByUser(userID string, page, pageSize int) ([]model.Checklist, int64, error) {
	var lists []model.Checklist
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Checklist{}).Where("user_id = ?", userID).Count(&total)
	err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Offset(offset).Limit(pageSize).Preload("Items").Find(&lists).Error
	return lists, total, err
}

// CreateChecklist 创建清单
func CreateChecklist(cl *model.Checklist) error {
	return database.DB.Create(cl).Error
}

// UpdateChecklistItem 更新清单条目勾选状态
func UpdateChecklistItem(id string, checked int) error {
	return database.DB.Model(&model.ChecklistItem{}).Where("id = ?", id).Update("checked", checked).Error
}

// GetChecklistDetail 获取单个备忘清单详情（含条目）
func GetChecklistDetail(id, userID string) (*model.Checklist, error) {
	var cl model.Checklist
	err := database.DB.Where("id = ? AND user_id = ?", id, userID).Preload("Items").First(&cl).Error
	return &cl, err
}

// UpdateChecklist 更新备忘清单（名称 + 替换条目）
func UpdateChecklist(id, userID string, name string, items []model.ChecklistItem) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 校验归属
		var cl model.Checklist
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&cl).Error; err != nil {
			return err
		}
		// 更新名称
		if name != "" {
			if err := tx.Model(&cl).Update("name", name).Error; err != nil {
				return err
			}
		}
		// 删除旧条目
		if err := tx.Where("checklist_id = ?", id).Delete(&model.ChecklistItem{}).Error; err != nil {
			return err
		}
		// 插入新条目
		for i := range items {
			items[i].ID = ""
			items[i].ChecklistID = id
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
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

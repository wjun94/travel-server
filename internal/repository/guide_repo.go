package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// ==================== Guide 攻略 ====================

// GetFeedGuides 获取已发布的攻略（瀑布流）
func GetFeedGuides(page, pageSize int, destination string) ([]model.Guide, int64, error) {
	var guides []model.Guide
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Guide{}).Where("status = ?", 1)
	if destination != "" {
		query = query.Where("destination LIKE ?", "%"+destination+"%")
	}
	query.Count(&total)
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&guides).Error
	return guides, total, err
}

// CreateGuide 创建攻略（仅基本信息）
func CreateGuide(guide *model.Guide) error {
	return database.DB.Create(guide).Error
}

// CreateGuideWithSections 事务创建攻略 + 板块
func CreateGuideWithSections(guide *model.Guide, sections []model.GuideSection) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(guide).Error; err != nil {
			return err
		}
		for i := range sections {
			sections[i].GuideID = guide.ID
			sections[i].SortOrder = i + 1
		}
		if len(sections) > 0 {
			if err := tx.Create(&sections).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetGuideByID 查询攻略详情（含板块）
func GetGuideByID(id uint) (*model.Guide, error) {
	var guide model.Guide
	err := database.DB.First(&guide, id).Error
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
func UpdateGuideStatus(id uint, status int) error {
	return database.DB.Model(&model.Guide{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateGuide 更新攻略基本信息（支持零值更新）
func UpdateGuide(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&model.Guide{}).Where("id = ?", id).Updates(updates).Error
}

// IncrementGuideViewCount 增加浏览量
func IncrementGuideViewCount(id uint) error {
	return database.DB.Model(&model.Guide{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// ==================== GuideSection 攻略板块 ====================

// GetSectionsByGuideID 获取攻略的所有板块（按排序序号排列）
func GetSectionsByGuideID(guideID uint) ([]model.GuideSection, error) {
	var sections []model.GuideSection
	err := database.DB.Where("guide_id = ?", guideID).
		Order("sort_order asc").Find(&sections).Error
	return sections, err
}

// CreateSection 创建攻略板块
func CreateSection(section *model.GuideSection) error {
	return database.DB.Create(section).Error
}

// UpdateSection 更新攻略板块
func UpdateSection(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&model.GuideSection{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteSection 删除攻略板块
func DeleteSection(id uint) error {
	return database.DB.Delete(&model.GuideSection{}, id).Error
}

// BatchCreateSections 批量创建板块
func BatchCreateSections(sections []model.GuideSection) error {
	return database.DB.Create(&sections).Error
}

// ReorderSections 重新排序板块
func ReorderSections(guideID uint, sectionIDs []uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for i, id := range sectionIDs {
			if err := tx.Model(&model.GuideSection{}).Where("id = ? AND guide_id = ?", id, guideID).
				Update("sort_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

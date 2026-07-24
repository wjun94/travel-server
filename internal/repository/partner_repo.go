package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// CreatePartner 发布搭子
func CreatePartner(p *model.Partner) error {
	return database.DB.Create(p).Error
}

// GetPartnerList 获取搭子列表
func GetPartnerList(page, pageSize int) ([]model.Partner, int64, error) {
	var list []model.Partner
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Partner{}).Count(&total)
	err := database.DB.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error
	return list, total, err
}

// GetPartnerByID 根据 ID 获取搭子
func GetPartnerByID(id string) (*model.Partner, error) {
	var p model.Partner
	err := database.DB.First(&p, "id = ?", id).Error
	return &p, err
}

// UpdatePartner 更新搭子信息
func UpdatePartner(p *model.Partner) error {
	return database.DB.Save(p).Error
}

// CreateApplication 创建搭子申请
func CreateApplication(app *model.PartnerApplication) error {
	return database.DB.Create(app).Error
}

// GetApplicationByID 获取申请详情
func GetApplicationByID(id string) (*model.PartnerApplication, error) {
	var app model.PartnerApplication
	err := database.DB.First(&app, "id = ?", id).Error
	return &app, err
}

// UpdateApplicationStatus 修改申请状态
func UpdateApplicationStatus(id string, status int) error {
	return database.DB.Model(&model.PartnerApplication{}).Where("id = ?", id).Update("status", status).Error
}

// GetUserAppliedPartnerIDs 获取当前用户已申请过的搭子ID集合（待审核+通过）
func GetUserAppliedPartnerIDs(userID string, partnerIDs []string) (map[string]bool, error) {
	var apps []model.PartnerApplication
	result := make(map[string]bool, len(partnerIDs))
	if len(partnerIDs) == 0 {
		return result, nil
	}
	if err := database.DB.Model(&model.PartnerApplication{}).
		Select("partner_id").
		Where("user_id = ? AND partner_id IN ? AND status IN (0, 1)", userID, partnerIDs).
		Find(&apps).Error; err != nil {
		return nil, err
	}
	for _, app := range apps {
		result[app.PartnerID] = true
	}
	return result, nil
}

// GetPartners 获取搭子列表（分页）
func GetPartners(page, pageSize int) ([]model.Partner, int64, error) {
	var list []model.Partner
	var total int64
	offset := (page - 1) * pageSize
	err := database.DB.Model(&model.Partner{}).Where("type = ?", 1).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = database.DB.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error
	return list, total, err
}

// GetMyPartners 获取当前用户创建的搭子列表
func GetMyPartners(userID string, page, pageSize int) ([]model.Partner, int64, error) {
	var list []model.Partner
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Partner{}).Where("user_id = ?", userID).Count(&total)
	err := database.DB.Where("user_id = ?", userID).Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error
	return list, total, err
}

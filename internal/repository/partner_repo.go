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
func GetPartnerByID(id uint) (*model.Partner, error) {
	var p model.Partner
	err := database.DB.First(&p, id).Error
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
func GetApplicationByID(id uint) (*model.PartnerApplication, error) {
	var app model.PartnerApplication
	err := database.DB.First(&app, id).Error
	return &app, err
}

// UpdateApplicationStatus 修改申请状态
func UpdateApplicationStatus(id uint, status int) error {
	return database.DB.Model(&model.PartnerApplication{}).Where("id = ?", id).Update("status", status).Error
}

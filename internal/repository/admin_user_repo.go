package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

func GetAdminUserByUsername(username string) (*model.AdminUser, error) {
	var user model.AdminUser
	err := database.DB.Preload("Role").Where("username = ? AND status = 1", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetAdminUserByID(id uint) (*model.AdminUser, error) {
	var user model.AdminUser
	err := database.DB.Preload("Role").First(&user, id).Error
	return &user, err
}

func ListAdminUsers(page, pageSize int) ([]model.AdminUser, int64, error) {
	var users []model.AdminUser
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.AdminUser{}).Count(&total)
	err := database.DB.Preload("Role").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func CreateAdminUser(user *model.AdminUser) error {
	return database.DB.Create(user).Error
}

func UpdateAdminUser(user *model.AdminUser) error {
	return database.DB.Save(user).Error
}

func DeleteAdminUser(id uint) error {
	return database.DB.Delete(&model.AdminUser{}, id).Error
}

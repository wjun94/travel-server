package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

func ListRoles(page, pageSize int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	database.DB.Model(&model.Role{}).Count(&total)
	err := database.DB.Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles).Error
	return roles, total, err
}

func GetRoleByID(id string) (*model.Role, error) {
	var role model.Role
	err := database.DB.First(&role, "id = ?", id).Error
	return &role, err
}

func CreateRole(role *model.Role) error {
	return database.DB.Create(role).Error
}

func UpdateRole(role *model.Role) error {
	return database.DB.Save(role).Error
}

func DeleteRole(id string) error {
	return database.DB.Where("id = ?", id).Delete(&model.Role{}).Error
}

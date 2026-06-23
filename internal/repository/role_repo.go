package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

func ListRoles() ([]model.Role, error) {
	var roles []model.Role
	err := database.DB.Find(&roles).Error
	return roles, err
}

func GetRoleByID(id uint) (*model.Role, error) {
	var role model.Role
	err := database.DB.First(&role, id).Error
	return &role, err
}

func CreateRole(role *model.Role) error {
	return database.DB.Create(role).Error
}

func UpdateRole(role *model.Role) error {
	return database.DB.Save(role).Error
}

func DeleteRole(id uint) error {
	return database.DB.Delete(&model.Role{}, id).Error
}

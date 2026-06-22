package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// GetUserByOpenID 根据 OpenID 查找用户
func GetUserByOpenID(openid string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("open_id = ?", openid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建新用户
func CreateUser(user *model.User) error {
	return database.DB.Create(user).Error
}

// GetUserByID 根据 ID 获取用户
func GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := database.DB.First(&user, id).Error
	return &user, err
}

// ListUsers 分页获取用户列表
func ListUsers(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.User{}).Count(&total)
	err := database.DB.Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// UpdateUserRole 更新用户角色
func UpdateUserRole(userID uint, role int) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).Update("role", role).Error
}

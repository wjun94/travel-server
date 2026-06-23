package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"travel-server/internal/model"
	"travel-server/internal/repository"
)

// AdminLogin 验证后台用户并返回用户信息
func AdminLogin(username, password string) (*model.AdminUser, error) {
	user, err := repository.GetAdminUserByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	return user, nil
}

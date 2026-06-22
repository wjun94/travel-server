package service

import (
	"errors"

	"gorm.io/gorm"

	"travel-server/internal/model"
	"travel-server/internal/repository"
)

// GetOrCreateUser 微信登录时查询或创建用户
func GetOrCreateUser(openid string) (*model.User, error) {
	user, err := repository.GetUserByOpenID(openid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，自动注册
			newUser := &model.User{OpenID: openid}
			if err := repository.CreateUser(newUser); err != nil {
				return nil, err
			}
			return newUser, nil
		}
		return nil, err
	}
	return user, nil
}

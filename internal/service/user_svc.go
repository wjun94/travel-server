package service

import (
	"errors"

	"gorm.io/gorm"

	"travel-server/internal/model"
	"travel-server/internal/repository"
)

// GetOrCreateUser 微信登录时查询或创建用户；unionid 为微信开放平台ID（可能为空），inviteCode 可选，仅新用户注册时绑定邀请关系
func GetOrCreateUser(openid, unionid, inviteCode string) (*model.User, error) {
	user, err := repository.GetUserByOpenID(openid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，自动注册
			newUser := &model.User{OpenID: openid, UnionID: unionid}
			if err := repository.CreateUser(newUser); err != nil {
				return nil, err
			}
			// 绑定邀请关系：新用户注册成功即视为邀请成功
			if inviteCode != "" {
				if inviter, err := repository.GetUserByInviteCode(inviteCode); err == nil {
					if err := repository.CreateInviteRecord(inviter.ID, newUser.ID); err != nil {
						return nil, err
					}
				}
			}
			return newUser, nil
		}
		return nil, err
	}
	// 老用户补全：unionid 之前为空且本次登录带回时，回填一次
	if user.UnionID == "" && unionid != "" {
		repository.UpdateUserUnionID(user.ID, unionid)
		user.UnionID = unionid
	}
	return user, nil
}

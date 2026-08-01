package repository

import (
	"time"

	"travel-server/internal/model"
	"travel-server/pkg/database"
	"travel-server/pkg/snowflake"
)

// EnsureInviteCode 确保用户有邀请码（老用户迁移兜底：为空时补发并持久化）
func EnsureInviteCode(userID string) (string, error) {
	user, err := GetUserByID(userID)
	if err != nil || user == nil {
		return "", err
	}
	if user.InviteCode != "" {
		return user.InviteCode, nil
	}
	code := snowflake.GenerateShortID(8)
	if err := database.DB.Model(&model.User{}).Where("id = ?", userID).Update("invite_code", code).Error; err != nil {
		return "", err
	}
	return code, nil
}

// GetUserByInviteCode 根据邀请码查询用户
func GetUserByInviteCode(inviteCode string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("invite_code = ?", inviteCode).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateInviteRecord 创建邀请记录（被邀请人ID唯一，重复绑定自动忽略）
func CreateInviteRecord(inviterID, inviteeID string) error {
	if inviterID == "" || inviteeID == "" || inviterID == inviteeID {
		return nil
	}
	rec := model.InviteRecord{
		InviterID: inviterID,
		InviteeID: inviteeID,
	}
	return database.DB.Create(&rec).Error
}

// CountTodayInviteSuccess 统计邀请人今日邀请成功人数（新用户注册数）
func CountTodayInviteSuccess(inviterID string) (int64, error) {
	var count int64
	start := time.Now().Truncate(24 * time.Hour)
	err := database.DB.Model(&model.InviteRecord{}).
		Where("inviter_id = ? AND created_at >= ?", inviterID, start).
		Count(&count).Error
	return count, err
}

// CountTotalInviteSuccess 统计邀请人累计邀请成功人数
func CountTotalInviteSuccess(inviterID string) (int64, error) {
	var count int64
	err := database.DB.Model(&model.InviteRecord{}).
		Where("inviter_id = ?", inviterID).
		Count(&count).Error
	return count, err
}

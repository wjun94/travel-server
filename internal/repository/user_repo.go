package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// UserProfileStats 用户个人主页统计数据
type UserProfileStats struct {
	ID            string `json:"id"`
	Nickname      string `json:"nickname"`
	AvatarURL     string `json:"avatarUrl"`
	GuideCount    int64  `json:"guideCount"`    // 已发布的攻略数
	TripCount     int64  `json:"tripCount"`     // 行程数
	FollowCount   int    `json:"followCount"`   // 关注数
	FollowerCount int    `json:"followerCount"` // 粉丝数
}

// GetUserProfileStats 获取用户主页统计
func GetUserProfileStats(userID string) (*UserProfileStats, error) {
	var user model.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	var guideCount int64
	database.DB.Model(&model.Guide{}).Where("user_id = ? AND status = 1", userID).Count(&guideCount)

	var tripCount int64
	database.DB.Model(&model.Trip{}).Where("user_id = ?", userID).Count(&tripCount)

	return &UserProfileStats{
		ID:            user.ID,
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		GuideCount:    guideCount,
		TripCount:     tripCount,
		FollowCount:   user.FollowCount,
		FollowerCount: user.FollowerCount,
	}, nil
}

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
func GetUserByID(id string) (*model.User, error) {
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
func UpdateUserRole(userID string, role int) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).Update("role", role).Error
}

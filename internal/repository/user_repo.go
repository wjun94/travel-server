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

// UserPublicProfile 他人个人主页（含获赞收藏和关注状态）
type UserPublicProfile struct {
	ID            string `json:"id"`
	Nickname      string `json:"nickname"`
	AvatarURL     string `json:"avatarUrl"`
	GuideCount    int64  `json:"guideCount"`    // 已发布的攻略数
	TripCount     int64  `json:"tripCount"`     // 行程数
	FollowCount   int    `json:"followCount"`   // 关注数
	FollowerCount int    `json:"followerCount"` // 粉丝数
	TotalLikes    int64  `json:"totalLikes"`    // 总获赞数
	TotalFavs     int64  `json:"totalFavs"`     // 总收藏数
	IsFollowed    bool   `json:"isFollowed"`    // 我是否已关注
	IsSelf        bool   `json:"isSelf"`        // 是否是自己
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

// GetUserPublicProfile 获取他人个人主页
func GetUserPublicProfile(userID, currentUserID string) (*UserPublicProfile, error) {
	var user model.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	var guideCount int64
	database.DB.Model(&model.Guide{}).Where("user_id = ? AND status = 1", userID).Count(&guideCount)

	var tripCount int64
	database.DB.Model(&model.Trip{}).Where("user_id = ?", userID).Count(&tripCount)

	// 获赞总数（用户所有已发布攻略的点赞数之和）
	var totalLikes int64
	database.DB.Model(&model.Guide{}).
		Select("COALESCE(SUM(like_count), 0)").
		Where("user_id = ? AND status = 1", userID).
		Scan(&totalLikes)

	// 收藏总数（用户内容被收藏的次数）
	var totalFavs int64
	database.DB.Raw(`SELECT COUNT(*) FROM favorites WHERE target_type IN ('guide','trip') AND target_id IN (
		SELECT id FROM guides WHERE user_id = ? AND status = 1
		UNION
		SELECT id FROM trips WHERE user_id = ?
	)`, userID, userID).Scan(&totalFavs)

	// 我是否已关注对方
	followStatus, _ := GetFollowStatus(currentUserID, userID)
	isFollowed := followStatus == 1 || followStatus == 2

	return &UserPublicProfile{
		ID:            user.ID,
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		GuideCount:    guideCount,
		TripCount:     tripCount,
		FollowCount:   user.FollowCount,
		FollowerCount: user.FollowerCount,
		TotalLikes:    totalLikes,
		TotalFavs:     totalFavs,
		IsFollowed:    isFollowed,
		IsSelf:        userID == currentUserID,
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

// GetUsersByIDs 批量查询用户信息
func GetUsersByIDs(ids []string) map[string]*model.User {
	result := make(map[string]*model.User, len(ids))
	if len(ids) == 0 {
		return result
	}
	var users []model.User
	database.DB.Where("id IN ?", ids).Find(&users)
	for i := range users {
		result[users[i].ID] = &users[i]
	}
	return result
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

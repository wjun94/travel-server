package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// UserProfileStats 用户个人主页统计数据
type UserProfileStats struct {
	ID                 string `json:"id"`
	Nickname           string `json:"nickname"`
	AvatarURL          string `json:"avatarUrl"`
	Gender             string `json:"gender"`             // 性别：unknown未知 male男 female女
	Role               int    `json:"role"`               // 0普通 1领队 2管理员
	GuideCount         int64  `json:"guideCount"`         // 已发布的攻略数
	TripCount          int64  `json:"tripCount"`          // 行程数
	PartnerCount       int64  `json:"partnerCount"`       // 搭子数（发布的）
	JoinedPartnerCount int64  `json:"joinedPartnerCount"` // 我参与的搭子数
	FollowCount        int    `json:"followCount"`        // 关注数
	FollowerCount      int    `json:"followerCount"`      // 粉丝数
	BlockCount         int64  `json:"blockCount"`         // 拉黑数
	TotalLikes         int64  `json:"totalLikes"`         // 总获赞数
	TotalFavs          int64  `json:"totalFavs"`          // 总收藏数
}

// UserPublicProfile 他人个人主页（含获赞收藏和关注状态）
type UserPublicProfile struct {
	ID                 string `json:"id"`
	Nickname           string `json:"nickname"`
	AvatarURL          string `json:"avatarUrl"`
	Gender             string `json:"gender"`             // 性别：unknown未知 male男 female女
	GuideCount         int64  `json:"guideCount"`         // 已发布的攻略数
	TripCount          int64  `json:"tripCount"`          // 行程数
	PartnerCount       int64  `json:"partnerCount"`       // 搭子数（发布的）
	JoinedPartnerCount int64  `json:"joinedPartnerCount"` // 我参与的搭子数
	FollowCount        int    `json:"followCount"`        // 关注数
	FollowerCount      int    `json:"followerCount"`      // 粉丝数
	TotalLikes         int64  `json:"totalLikes"`         // 总获赞数
	TotalFavs          int64  `json:"totalFavs"`          // 总收藏数
	IsFollowed         bool   `json:"isFollowed"`         // 我是否已关注
	IsBlocked          bool   `json:"isBlocked"`          // 我是否已拉黑对方
	IsSelf             bool   `json:"isSelf"`             // 是否是自己
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

	var partnerCount int64
	database.DB.Model(&model.Partner{}).Where("user_id = ?", userID).Count(&partnerCount)

	joinedPartnerCount, _ := CountJoinedPartners(userID)

	var blockCount int64
	database.DB.Model(&model.Follow{}).Where("user_id = ? AND status = 1", userID).Count(&blockCount)

	// 获赞总数（用户所有攻略/行程/搭子的点赞数之和）
	var totalLikes int64
	database.DB.Raw(`SELECT COALESCE(
		(SELECT SUM(like_count) FROM guides WHERE user_id = ?),
		0) + COALESCE(
		(SELECT SUM(like_count) FROM trips WHERE user_id = ?),
		0) + COALESCE(
		(SELECT SUM(like_count) FROM partners WHERE user_id = ?),
		0)`, userID, userID, userID).Scan(&totalLikes)

	// 收藏总数（用户攻略/行程/搭子被收藏的次数）
	var totalFavs int64
	database.DB.Raw(`SELECT COUNT(*) FROM favorites WHERE action = 2 AND target_type IN ('guide','trip','partner') AND target_id IN (
		SELECT id FROM guides WHERE user_id = ?
		UNION
		SELECT id FROM trips WHERE user_id = ?
		UNION
		SELECT id FROM partners WHERE user_id = ?
	)`, userID, userID, userID).Scan(&totalFavs)

	return &UserProfileStats{
		ID:                 user.ID,
		Nickname:           user.Nickname,
		AvatarURL:          user.AvatarURL,
		Gender:             user.Gender,
		Role:               user.Role,
		GuideCount:         guideCount,
		TripCount:          tripCount,
		PartnerCount:       partnerCount,
		JoinedPartnerCount: joinedPartnerCount,
		FollowCount:        user.FollowCount,
		FollowerCount:      user.FollowerCount,
		BlockCount:         blockCount,
		TotalLikes:         totalLikes,
		TotalFavs:          totalFavs,
	}, nil
}

// UpdateUserUnionID 回填用户微信开放平台ID（unionid）
func UpdateUserUnionID(userID, unionid string) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).Update("union_id", unionid).Error
}

// UpdateUserPhone 更新用户手机号
func UpdateUserPhone(userID, phone string) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).Update("phone", phone).Error
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

	var partnerCount int64
	database.DB.Model(&model.Partner{}).Where("user_id = ?", userID).Count(&partnerCount)

	joinedPartnerCount, _ := CountJoinedPartners(userID)

	// 获赞总数（用户所有攻略/行程/搭子的点赞数之和）
	var totalLikes int64
	database.DB.Raw(`SELECT COALESCE(
		(SELECT SUM(like_count) FROM guides WHERE user_id = ?),
		0) + COALESCE(
		(SELECT SUM(like_count) FROM trips WHERE user_id = ?),
		0) + COALESCE(
		(SELECT SUM(like_count) FROM partners WHERE user_id = ?),
		0)`, userID, userID, userID).Scan(&totalLikes)

	// 收藏总数（用户攻略/行程/搭子被收藏的次数）
	var totalFavs int64
	database.DB.Raw(`SELECT COUNT(*) FROM favorites WHERE action = 2 AND target_type IN ('guide','trip','partner') AND target_id IN (
		SELECT id FROM guides WHERE user_id = ?
		UNION
		SELECT id FROM trips WHERE user_id = ?
		UNION
		SELECT id FROM partners WHERE user_id = ?
	)`, userID, userID, userID).Scan(&totalFavs)

	// 我是否已关注对方
	followStatus, _ := GetFollowStatus(currentUserID, userID)
	isFollowed := followStatus == 1 || followStatus == 2

	// 我是否已拉黑对方
	var blockedCount int64
	database.DB.Model(&model.Follow{}).
		Where("user_id = ? AND follower_id = ? AND status = 1", currentUserID, userID).
		Count(&blockedCount)

	return &UserPublicProfile{
		ID:                 user.ID,
		Nickname:           user.Nickname,
		AvatarURL:          user.AvatarURL,
		Gender:             user.Gender,
		GuideCount:         guideCount,
		TripCount:          tripCount,
		PartnerCount:       partnerCount,
		JoinedPartnerCount: joinedPartnerCount,
		FollowCount:        user.FollowCount,
		FollowerCount:      user.FollowerCount,
		TotalLikes:         totalLikes,
		TotalFavs:          totalFavs,
		IsFollowed:         isFollowed,
		IsBlocked:          blockedCount > 0,
		IsSelf:             userID == currentUserID,
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

// ListUsers 分页获取用户列表（支持昵称关键词筛选）
func ListUsers(page, pageSize int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.User{})
	if keyword != "" {
		query = query.Where("nickname LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&users).Error
	return users, total, err
}

// UpdateUserRole 更新用户角色
func UpdateUserRole(userID string, role int) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).Update("role", role).Error
}

// ListAllUserIDs 获取全部用户ID（系统消息-全部用户）
func ListAllUserIDs() ([]string, error) {
	var ids []string
	err := database.DB.Model(&model.User{}).Pluck("id", &ids).Error
	return ids, err
}

// ListUserIDsByRole 获取指定角色的用户ID列表（系统消息-用户分组）
func ListUserIDsByRole(role int) ([]string, error) {
	var ids []string
	err := database.DB.Model(&model.User{}).Where("role = ?", role).Pluck("id", &ids).Error
	return ids, err
}

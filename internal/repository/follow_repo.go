package repository

import (
	"errors"
	"fmt"

	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// FollowItem 关注列表项（含用户信息和互关状态）
type FollowItem struct {
	UserID    string `json:"userId"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
	IsMutual  bool   `json:"isMutual"` // 是否互关
	IsFollow  bool   `json:"isFollow"` // 我是否关注了对方（粉丝列表用）
}

// ==================== 1. 关注用户 ====================

func FollowUser(userID, followerID string) error {
	if userID == followerID {
		return errors.New("不能关注自己")
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 检查是否已在对方拉黑名单（对方拉黑了我）
		var iBlocked int64
		tx.Model(&model.Follow{}).
			Where("user_id = ? AND follower_id = ? AND status = 1", userID, followerID).
			Count(&iBlocked)
		if iBlocked > 0 {
			return errors.New("已被对方拉黑，无法关注")
		}
		// 检查我是否拉黑了对方
		var theyBlocked int64
		tx.Model(&model.Follow{}).
			Where("user_id = ? AND follower_id = ? AND status = 1", followerID, userID).
			Count(&theyBlocked)
		if theyBlocked > 0 {
			return errors.New("对方在你的黑名单中，请先取消拉黑")
		}
		// 查重
		var count int64
		tx.Model(&model.Follow{}).
			Where("user_id = ? AND follower_id = ? AND status = 0", userID, followerID).
			Count(&count)
		if count > 0 {
			return nil // 幂等
		}
		// 创建关注关系
		if err := tx.Create(&model.Follow{
			UserID:     userID,
			FollowerID: followerID,
		}).Error; err != nil {
			return err
		}
		// 操作人关注数 +1
		if err := tx.Model(&model.User{}).Where("id = ?", followerID).
			UpdateColumn("follow_count", gorm.Expr("follow_count + 1")).Error; err != nil {
			return err
		}
		// 被关注人粉丝数 +1
		if err := tx.Model(&model.User{}).Where("id = ?", userID).
			UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).Error; err != nil {
			return err
		}
		// 发送新粉丝通知
		return createFollowNotification(followerID, userID)
	})
}

// createFollowNotification 创建新粉丝通知
func createFollowNotification(followerID, toUserID string) error {
	var follower model.User
	if err := database.DB.Select("nickname").First(&follower, "id = ?", followerID).Error; err != nil {
		return nil
	}
	return database.DB.Create(&model.Message{
		FromUserID: followerID,
		ToUserID:   toUserID,
		Content:    fmt.Sprintf("%s 关注了你", follower.Nickname),
		Type:       2,
	}).Error
}

// ==================== 2. 取消关注 ====================

func UnfollowUser(userID, followerID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND follower_id = ? AND status = 0", userID, followerID).
			Delete(&model.Follow{})
		if result.RowsAffected == 0 {
			return errors.New("未关注该用户")
		}
		// 操作人关注数 -1
		if err := tx.Model(&model.User{}).Where("id = ?", followerID).
			UpdateColumn("follow_count", gorm.Expr("GREATEST(follow_count - 1, 0)")).Error; err != nil {
			return err
		}
		// 被关注人粉丝数 -1
		return tx.Model(&model.User{}).Where("id = ?", userID).
			UpdateColumn("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)")).Error
	})
}

// ==================== 3. 我的关注列表 ====================

func GetMyFollowingList(followerID string, page, pageSize int) ([]FollowItem, int64, error) {
	var total int64
	query := database.DB.Model(&model.Follow{}).
		Where("follower_id = ? AND status = 0", followerID)
	query.Count(&total)

	var follows []model.Follow
	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		return nil, total, err
	}

	return buildFollowItems(follows, "user_id", followerID, true, false)
}

// ==================== 4. 我的粉丝列表 ====================

func GetMyFollowerList(userID string, page, pageSize int) ([]FollowItem, int64, error) {
	var total int64
	query := database.DB.Model(&model.Follow{}).
		Where("user_id = ? AND status = 0", userID)
	query.Count(&total)

	var follows []model.Follow
	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		return nil, total, err
	}

	return buildFollowItems(follows, "follower_id", userID, false, true)
}

// ==================== 5. 他人关注列表 ====================

func GetUserFollowingList(targetUserID, currentUserID string, page, pageSize int) ([]FollowItem, int64, error) {
	if !canViewFollow(targetUserID, currentUserID) {
		return []FollowItem{}, 0, nil
	}

	var total int64
	query := database.DB.Model(&model.Follow{}).
		Where("follower_id = ? AND status = 0", targetUserID)
	query.Count(&total)

	var follows []model.Follow
	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		return nil, total, err
	}

	return buildFollowItems(follows, "user_id", currentUserID, true, false)
}

// ==================== 6. 他人粉丝列表 ====================

func GetUserFollowerList(targetUserID, currentUserID string, page, pageSize int) ([]FollowItem, int64, error) {
	if !canViewFollow(targetUserID, currentUserID) {
		return []FollowItem{}, 0, nil
	}

	var total int64
	query := database.DB.Model(&model.Follow{}).
		Where("user_id = ? AND status = 0", targetUserID)
	query.Count(&total)

	var follows []model.Follow
	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		return nil, total, err
	}

	return buildFollowItems(follows, "follower_id", currentUserID, false, true)
}

// canViewFollow 校验是否有权限查看对方的关注/粉丝（对方拉黑了我则无权）
func canViewFollow(targetUserID, currentUserID string) bool {
	if targetUserID == currentUserID {
		return true
	}
	var blocked int64
	database.DB.Model(&model.Follow{}).
		Where("user_id = ? AND follower_id = ? AND status = 1", targetUserID, currentUserID).
		Count(&blocked)
	return blocked == 0
}

// ==================== 7. 关系状态 ====================

func GetFollowStatus(currentUserID, targetUserID string) (int, error) {
	if currentUserID == targetUserID {
		return 0, nil
	}
	// 是否存在拉黑关系
	var blocked int64
	database.DB.Model(&model.Follow{}).
		Where("(user_id = ? AND follower_id = ? AND status = 1) OR (user_id = ? AND follower_id = ? AND status = 1)",
			targetUserID, currentUserID, currentUserID, targetUserID).
		Count(&blocked)
	if blocked > 0 {
		return 3, nil
	}
	// 我是否关注了对方
	var iFollow int64
	database.DB.Model(&model.Follow{}).
		Where("user_id = ? AND follower_id = ? AND status = 0", targetUserID, currentUserID).
		Count(&iFollow)
	// 对方是否关注了我
	var theyFollow int64
	database.DB.Model(&model.Follow{}).
		Where("user_id = ? AND follower_id = ? AND status = 0", currentUserID, targetUserID).
		Count(&theyFollow)

	if iFollow > 0 && theyFollow > 0 {
		return 2, nil
	}
	if iFollow > 0 {
		return 1, nil
	}
	return 0, nil
}

// GetFollowStatusMap 批量查询当前用户对多个目标用户的关注状态
// 返回 map[targetUserID]status，status 含义同 GetFollowStatus
func GetFollowStatusMap(currentUserID string, targetUserIDs []string) map[string]int {
	result := make(map[string]int, len(targetUserIDs))
	for _, id := range targetUserIDs {
		result[id] = 0
	}
	if len(targetUserIDs) == 0 || currentUserID == "" {
		return result
	}

	// 查所有关注记录：我关注了谁
	var follows []struct {
		UserID string
	}
	database.DB.Model(&model.Follow{}).
		Where("user_id IN ? AND follower_id = ? AND status = 0", targetUserIDs, currentUserID).
		Find(&follows)
	for _, f := range follows {
		result[f.UserID] = 1
	}

	// 查谁关注了我
	var followers []struct {
		UserID string
	}
	database.DB.Model(&model.Follow{}).
		Where("user_id = ? AND follower_id IN ? AND status = 0", currentUserID, targetUserIDs).
		Find(&followers)
	for _, f := range followers {
		if result[f.UserID] == 1 {
			result[f.UserID] = 2 // 互相关注
		}
	}

	// 查拉黑关系
	var blocks []struct {
		UserID string
	}
	database.DB.Model(&model.Follow{}).
		Where("(user_id IN ? AND follower_id = ? AND status = 1) OR (user_id = ? AND follower_id IN ? AND status = 1)",
			targetUserIDs, currentUserID, currentUserID, targetUserIDs).
		Find(&blocks)
	for _, b := range blocks {
		result[b.UserID] = 3
	}

	return result
}

// ==================== 8. 我的关注粉丝总数 ====================

func GetMyCounts(userID string) (int, int, error) {
	var user model.User
	if err := database.DB.Select("follow_count", "follower_count").
		First(&user, "id = ?", userID).Error; err != nil {
		return 0, 0, err
	}
	return user.FollowCount, user.FollowerCount, nil
}

// ==================== 9. 他人关注粉丝总数 ====================

func GetUserCounts(targetUserID, currentUserID string) (int, int, error) {
	if !canViewFollow(targetUserID, currentUserID) {
		return 0, 0, nil
	}
	var user model.User
	if err := database.DB.Select("follow_count", "follower_count").
		First(&user, "id = ?", targetUserID).Error; err != nil {
		return 0, 0, err
	}
	return user.FollowCount, user.FollowerCount, nil
}

// ==================== 10. 移除粉丝 ====================

func RemoveFollower(userID, followerID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND follower_id = ? AND status = 0", userID, followerID).
			Delete(&model.Follow{})
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		// 粉丝（关注者）关注数 -1
		if err := tx.Model(&model.User{}).Where("id = ?", followerID).
			UpdateColumn("follow_count", gorm.Expr("GREATEST(follow_count - 1, 0)")).Error; err != nil {
			return err
		}
		// 本人粉丝数 -1
		return tx.Model(&model.User{}).Where("id = ?", userID).
			UpdateColumn("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)")).Error
	})
}

// ==================== 工具：构建列表项 ====================

// buildFollowItems 从 Follow 记录批量构建列表，附加用户信息和关系标注
// targetColumn: "user_id" 表示目标用户在 UserID 字段，"follower_id" 表示在 FollowerID 字段
// currentUserID: 当前登录用户
// needMutual: 是否需要标注 isMutual（对方是否关注了我）
// needFollow: 是否需要标注 isFollow（我是否关注了对方）
func buildFollowItems(follows []model.Follow, targetColumn, currentUserID string, needMutual, needFollow bool) ([]FollowItem, int64, error) {
	total := int64(len(follows))
	items := make([]FollowItem, 0, len(follows))

	// 收集目标用户ID
	var targetIDs []string
	for _, f := range follows {
		if targetColumn == "user_id" {
			targetIDs = append(targetIDs, f.UserID)
		} else {
			targetIDs = append(targetIDs, f.FollowerID)
		}
	}
	if len(targetIDs) == 0 {
		return items, 0, nil
	}

	// 批量查询用户信息
	var users []model.User
	database.DB.Where("id IN ?", targetIDs).Find(&users)
	userMap := make(map[string]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// 批量查询互关状态（如果需要）
	var mutualMap map[string]bool
	if needMutual {
		mutualMap = getRelationMap(currentUserID, targetIDs, "user_id")
	}
	// 批量查询关注状态（如果需要）
	var followMap map[string]bool
	if needFollow {
		followMap = getRelationMap(currentUserID, targetIDs, "follower_id")
	}

	// 过滤掉被我拉黑的用户
	blockedIDs := getBlockedUserIDs(currentUserID)

	for _, f := range follows {
		var uid string
		if targetColumn == "user_id" {
			uid = f.UserID
		} else {
			uid = f.FollowerID
		}
		// 跳过拉黑用户
		if blockedIDs[uid] {
			continue
		}
		u, ok := userMap[uid]
		if !ok {
			continue
		}
		item := FollowItem{
			UserID:    uid,
			Nickname:  u.Nickname,
			AvatarURL: u.AvatarURL,
		}
		if needMutual {
			item.IsMutual = mutualMap[uid]
		}
		if needFollow {
			item.IsFollow = followMap[uid]
		}
		items = append(items, item)
	}
	return items, total, nil
}

// getRelationMap 批量查询 currentUserID 与 targetIDs 的关系
// direction="user_id" → 查 currentUserID 是否在 target 的粉丝列表中（对方关注了我）
// direction="follower_id" → 查 currentUserID 是否关注了 target
func getRelationMap(currentUserID string, targetIDs []string, direction string) map[string]bool {
	result := make(map[string]bool, len(targetIDs))
	if direction == "user_id" {
		// 查 target 是否关注了我：Follow where user_id=currentUserID AND follower_id=target
		var follows []model.Follow
		database.DB.Where("user_id = ? AND follower_id IN ? AND status = 0", currentUserID, targetIDs).
			Find(&follows)
		for _, f := range follows {
			result[f.FollowerID] = true
		}
	} else {
		// 查我是否关注了 target：Follow where user_id=target AND follower_id=currentUserID
		var follows []model.Follow
		database.DB.Where("user_id IN ? AND follower_id = ? AND status = 0", targetIDs, currentUserID).
			Find(&follows)
		for _, f := range follows {
			result[f.UserID] = true
		}
	}
	return result
}

// getBlockedUserIDs 获取当前用户拉黑了哪些用户
func getBlockedUserIDs(userID string) map[string]bool {
	var blocks []model.Follow
	database.DB.Where("user_id = ? AND status = 1", userID).Find(&blocks)
	result := make(map[string]bool, len(blocks))
	for _, b := range blocks {
		result[b.FollowerID] = true
	}
	return result
}

// ==================== 11. 拉黑用户 ====================

func BlockUser(userID, blockedUserID string) error {
	if userID == blockedUserID {
		return errors.New("不能拉黑自己")
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 查重
		var count int64
		tx.Model(&model.Follow{}).
			Where("user_id = ? AND follower_id = ? AND status = 1", userID, blockedUserID).
			Count(&count)
		if count > 0 {
			return nil // 幂等
		}
		// 创建拉黑记录
		if err := tx.Create(&model.Follow{
			UserID:     userID,
			FollowerID: blockedUserID,
			Status:     1,
		}).Error; err != nil {
			return err
		}
		// 若我关注了对方 → 取消关注（我的关注数-1，对方粉丝数-1）
		res1 := tx.Where("user_id = ? AND follower_id = ? AND status = 0", blockedUserID, userID).
			Delete(&model.Follow{})
		if res1.RowsAffected > 0 {
			tx.Model(&model.User{}).Where("id = ?", userID).
				UpdateColumn("follow_count", gorm.Expr("GREATEST(follow_count - 1, 0)"))
			tx.Model(&model.User{}).Where("id = ?", blockedUserID).
				UpdateColumn("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)"))
		}
		// 若对方关注了我 → 取消关注（对方关注数-1，我的粉丝数-1）
		res2 := tx.Where("user_id = ? AND follower_id = ? AND status = 0", userID, blockedUserID).
			Delete(&model.Follow{})
		if res2.RowsAffected > 0 {
			tx.Model(&model.User{}).Where("id = ?", blockedUserID).
				UpdateColumn("follow_count", gorm.Expr("GREATEST(follow_count - 1, 0)"))
			tx.Model(&model.User{}).Where("id = ?", userID).
				UpdateColumn("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)"))
		}
		return nil
	})
}

// ==================== 12. 解除拉黑 ====================

func UnblockUser(userID, blockedUserID string) error {
	result := database.DB.Where("user_id = ? AND follower_id = ? AND status = 1", userID, blockedUserID).
		Delete(&model.Follow{})
	if result.RowsAffected == 0 {
		return errors.New("未拉黑该用户")
	}
	return nil
}

// ==================== 13. 我的拉黑名单 ====================

type BlacklistItem struct {
	UserID    string `json:"userId"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
}

func GetMyBlacklist(userID string, page, pageSize int) ([]BlacklistItem, int64, error) {
	var total int64
	query := database.DB.Model(&model.Follow{}).
		Where("user_id = ? AND status = 1", userID)
	query.Count(&total)

	var blocks []model.Follow
	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&blocks).Error; err != nil {
		return nil, total, err
	}

	// 查询被拉黑用户的个人信息
	var blockedIDs []string
	for _, b := range blocks {
		blockedIDs = append(blockedIDs, b.FollowerID)
	}
	if len(blockedIDs) == 0 {
		return []BlacklistItem{}, total, nil
	}
	var users []model.User
	database.DB.Where("id IN ?", blockedIDs).Find(&users)
	userMap := make(map[string]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	items := make([]BlacklistItem, 0, len(blocks))
	for _, b := range blocks {
		u, ok := userMap[b.FollowerID]
		if !ok {
			continue
		}
		items = append(items, BlacklistItem{
			UserID:    u.ID,
			Nickname:  u.Nickname,
			AvatarURL: u.AvatarURL,
		})
	}
	return items, total, nil
}

// ==================== 14. 校验是否被对方拉黑 ====================

func IsBlockedByUser(currentUserID, targetUserID string) (bool, error) {
	var count int64
	err := database.DB.Model(&model.Follow{}).
		Where("user_id = ? AND follower_id = ? AND status = 1", targetUserID, currentUserID).
		Count(&count).Error
	return count > 0, err
}

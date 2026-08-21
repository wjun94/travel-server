package repository

import (
	"errors"

	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// CreateComment 创建评论
func CreateComment(comment *model.Comment) error {
	return database.DB.Create(comment).Error
}

// GetCommentsByTarget 获取目标评论列表（含回复嵌套）
func GetCommentsByTarget(targetType string, targetID string, page, pageSize int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Comment{}).
		Where("target_type = ? AND target_id = ? AND parent_id IS NULL", targetType, targetID).
		Count(&total)
	err := database.DB.Where("target_type = ? AND target_id = ? AND parent_id IS NULL", targetType, targetID).
		Order("created_at desc").Offset(offset).Limit(pageSize).Preload("User").Find(&comments).Error
	return comments, total, err
}

// GetRepliesByParentID 获取子回复（递归查全部后代）
func GetRepliesByParentID(parentID string) ([]model.Comment, error) {
	allReplies := make([]model.Comment, 0)
	currentIDs := []string{parentID}
	for len(currentIDs) > 0 {
		var replies []model.Comment
		err := database.DB.Where("parent_id IN ?", currentIDs).
			Order("created_at asc").Preload("User").Find(&replies).Error
		if err != nil {
			return nil, err
		}
		if len(replies) == 0 {
			break
		}
		allReplies = append(allReplies, replies...)
		currentIDs = make([]string, len(replies))
		for i, r := range replies {
			currentIDs[i] = r.ID
		}
	}
	return allReplies, nil
}

// GetReplyCounts 批量查询多条评论的回复数（递归查全部后代，与 GetRepliesByParentID 保持一致）
func GetReplyCounts(parentIDs []string) map[string]int64 {
	result := make(map[string]int64)
	if len(parentIDs) == 0 {
		return result
	}
	for _, pid := range parentIDs {
		result[pid] = 0
	}

	// childID → 最初的顶级评论ID
	childToTop := make(map[string]string)

	currentIDs := parentIDs
	for len(currentIDs) > 0 {
		type replyRow struct {
			ParentID string
			ID       string
		}
		var direct []replyRow
		database.DB.Model(&model.Comment{}).
			Select("parent_id, id").
			Where("parent_id IN ?", currentIDs).
			Find(&direct)
		if len(direct) == 0 {
			break
		}

		nextIDs := make([]string, 0, len(direct))
		for _, r := range direct {
			// 找出这条回复归属的顶级评论
			topID, ok := childToTop[r.ParentID]
			if !ok {
				topID = r.ParentID // ParentID 本身就在 parentIDs 中
			}
			childToTop[r.ID] = topID
			result[topID]++
			nextIDs = append(nextIDs, r.ID)
		}
		currentIDs = nextIDs
	}

	return result
}

// DeleteComment 删除评论（评论作者本人或帖子作者均可，级联删除子回复）
func DeleteComment(id, userID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 校验归属：先按评论作者查
		var c model.Comment
		if err := tx.Where("id = ?", id).First(&c).Error; err != nil {
			return err
		}
		if c.UserID != userID {
			// 评论作者非本人时，允许帖子（攻略/行程/搭子）作者删除
			authorID, err := GetTargetAuthorID(c.TargetType, c.TargetID)
			if err != nil || authorID != userID {
				return ErrCommentNotFound
			}
		}
		// 级联删除子回复
		if err := tx.Where("parent_id = ?", id).Delete(&model.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&c).Error
	})
}

// GetTargetAuthorID 获取帖子作者ID（guide/trip/partner）
func GetTargetAuthorID(targetType, targetID string) (string, error) {
	switch targetType {
	case "guide":
		var g model.Guide
		if err := database.DB.Select("user_id").Where("id = ?", targetID).First(&g).Error; err != nil {
			return "", err
		}
		return g.UserID, nil
	case "trip":
		var t model.Trip
		if err := database.DB.Select("user_id").Where("id = ?", targetID).First(&t).Error; err != nil {
			return "", err
		}
		return t.UserID, nil
	case "partner":
		var p model.Partner
		if err := database.DB.Select("user_id").Where("id = ?", targetID).First(&p).Error; err != nil {
			return "", err
		}
		return p.UserID, nil
	}
	return "", ErrCommentNotFound
}

// ErrCommentNotFound 评论不存在或无权删除
var ErrCommentNotFound = errors.New("评论不存在")

// IncrementCommentLikeCount 点赞评论
func IncrementCommentLikeCount(id string) error {
	return database.DB.Model(&model.Comment{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

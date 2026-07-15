package repository

import (
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

// DeleteComment 删除评论（校验用户归属，级联删除子回复）
func DeleteComment(id, userID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 校验归属
		var c model.Comment
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&c).Error; err != nil {
			return err
		}
		// 级联删除子回复
		if err := tx.Where("parent_id = ?", id).Delete(&model.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&c).Error
	})
}

// IncrementCommentLikeCount 点赞评论
func IncrementCommentLikeCount(id string) error {
	return database.DB.Model(&model.Comment{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

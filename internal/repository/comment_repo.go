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
func GetCommentsByTarget(targetType string, targetID uint, page, pageSize int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Comment{}).
		Where("target_type = ? AND target_id = ? AND parent_id IS NULL", targetType, targetID).
		Count(&total)
	err := database.DB.Where("target_type = ? AND target_id = ? AND parent_id IS NULL", targetType, targetID).
		Order("created_at desc").Offset(offset).Limit(pageSize).Find(&comments).Error
	return comments, total, err
}

// GetRepliesByParentID 获取子回复
func GetRepliesByParentID(parentID uint) ([]model.Comment, error) {
	var replies []model.Comment
	err := database.DB.Where("parent_id = ?", parentID).
		Order("created_at asc").Find(&replies).Error
	return replies, err
}

// DeleteComment 删除评论
func DeleteComment(id uint) error {
	return database.DB.Delete(&model.Comment{}, id).Error
}

// IncrementCommentLikeCount 点赞评论
func IncrementCommentLikeCount(id uint) error {
	return database.DB.Model(&model.Comment{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

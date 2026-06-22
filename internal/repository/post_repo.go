package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// GetFeedPosts 获取已发布的攻略（瀑布流）
func GetFeedPosts(page, pageSize int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Post{}).Where("status = ?", 1).Count(&total)
	err := database.DB.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&posts).Error
	return posts, total, err
}

// CreatePost 发布攻略
func CreatePost(post *model.Post) error {
	return database.DB.Create(post).Error
}

// GetPostByID 查询攻略详情
func GetPostByID(id uint) (*model.Post, error) {
	var post model.Post
	err := database.DB.First(&post, id).Error
	return &post, err
}

// ListPosts 后台攻略列表（所有状态）
func ListPosts(page, pageSize int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Post{}).Count(&total)
	err := database.DB.Offset(offset).Limit(pageSize).Find(&posts).Error
	return posts, total, err
}

// UpdatePostStatus 审核攻略（修改状态）
func UpdatePostStatus(id uint, status int) error {
	return database.DB.Model(&model.Post{}).Where("id = ?", id).Update("status", status).Error
}

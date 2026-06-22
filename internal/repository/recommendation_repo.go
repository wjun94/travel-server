package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// CreateRecommendation 保存推荐内容
func CreateRecommendation(rec *model.Recommendation) error {
	return database.DB.Create(rec).Error
}

// GetRecommendations 获取推荐列表（可按城市筛选）
func GetRecommendations(city string) ([]model.Recommendation, error) {
	var list []model.Recommendation
	db := database.DB.Model(&model.Recommendation{})
	if city != "" {
		db = db.Where("city = ?", city)
	}
	err := db.Order("created_at desc").Find(&list).Error
	return list, err
}

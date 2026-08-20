package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// CreateAiGenerateLog 记录AI生成点击日志（status 0已点击）
func CreateAiGenerateLog(userID, logType string) (*model.AiGenerateLog, error) {
	l := model.AiGenerateLog{
		UserID: userID,
		Type:   logType,
		Status: 0,
	}
	if err := database.DB.Create(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

// UpdateAiGenerateLogStatus 更新AI生成日志状态（1生成成功 2生成失败）
func UpdateAiGenerateLogStatus(id string, status int) error {
	return database.DB.Model(&model.AiGenerateLog{}).Where("id = ?", id).Update("status", status).Error
}

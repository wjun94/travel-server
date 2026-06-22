package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// GetAccountsByTrip 获取某行程的记账记录
func GetAccountsByTrip(tripID, userID uint) ([]model.Accounting, error) {
	var accounts []model.Accounting
	err := database.DB.Where("trip_id = ? AND user_id = ?", tripID, userID).Find(&accounts).Error
	return accounts, err
}

// CreateAccount 添加记账条目
func CreateAccount(acc *model.Accounting) error {
	return database.DB.Create(acc).Error
}

package service

import (
	"encoding/json"

	"travel-server/internal/model"
	"travel-server/internal/repository"
)

// ApplyTripEdit 处理 WebSocket 传来的行程编辑，更新数据库
func ApplyTripEdit(tripID uint, editMsg map[string]interface{}) error {
	trip, err := repository.GetTripByID(tripID)
	if err != nil {
		return err
	}
	// 只更新 daily_plans 字段
	if plans, ok := editMsg["daily_plans"]; ok {
		b, _ := json.Marshal(plans)
		trip.DailyPlans = model.JSONString(b)
	}
	return repository.UpdateTrip(trip)
}

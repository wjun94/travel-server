package service

import (
	"travel-server/internal/repository"
)

// ApplyTripEdit 处理 WebSocket 传来的行程编辑，更新数据库
func ApplyTripEdit(tripID string, editMsg map[string]interface{}) error {
	_, err := repository.GetTripByID(tripID)
	if err != nil {
		return err
	}
	// 过滤安全字段后更新
	delete(editMsg, "id")
	delete(editMsg, "user_id")
	delete(editMsg, "created_at")
	return repository.UpdateTrip(tripID, editMsg)
}

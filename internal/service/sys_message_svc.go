package service

import (
	"encoding/json"
	"log"
	"time"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/internal/ws"
	"travel-server/pkg/database"
)

// 用户分组 → 角色映射（与 User.Role 对应：0普通 1领队 2管理员）
var userGroupRoleMap = map[string]int{
	"normal": 0,
	"leader": 1,
	"admin":  2,
}

// ResolveTargetUserIDs 解析系统消息的接收人群，返回目标用户ID列表
func ResolveTargetUserIDs(msg *model.SysMessage) ([]string, error) {
	switch msg.TargetType {
	case "all":
		return repository.ListAllUserIDs()
	case "group":
		role, ok := userGroupRoleMap[msg.TargetGroup]
		if !ok {
			return nil, nil
		}
		return repository.ListUserIDsByRole(role)
	case "users":
		var ids []string
		if err := json.Unmarshal([]byte(msg.TargetUserIDs), &ids); err != nil {
			return nil, err
		}
		return ids, nil
	}
	return nil, nil
}

// DeliverSysMessage 将系统消息以通知（type=4）形式下发给指定用户，返回实际送达人数
func DeliverSysMessage(msg *model.SysMessage, userIDs []string) int {
	if len(userIDs) == 0 {
		return 0
	}
	now := time.Now()
	notifications := make([]model.Notification, 0, len(userIDs))
	for _, uid := range userIDs {
		notifications = append(notifications, model.Notification{
			UserID:    uid,
			Type:      4, // 系统通知
			RelatedID: msg.ID,
			Content:   msg.Content,
			Title:     msg.Title,
			LinkURL:   msg.LinkURL,
			IsRead:    0,
			CreatedAt: now,
		})
	}
	if err := database.DB.Create(&notifications).Error; err != nil {
		log.Printf("系统消息 %s 下发失败: %v", msg.ID, err)
		return 0
	}

	// 实时推送：通知在线的目标用户刷新消息中心
	payload := map[string]interface{}{
		"action":  "new_notification",
		"type":    4, // 系统通知
		"title":   msg.Title,
		"content": msg.Content,
	}
	for _, uid := range userIDs {
		ws.WsHub.PushToUser(uid, payload)
	}
	return len(notifications)
}

// SendSysMessage 发送一条系统消息：解析人群 → 下发通知 → 更新消息状态，返回实际送达人数
func SendSysMessage(msg *model.SysMessage) (int, error) {
	userIDs, err := ResolveTargetUserIDs(msg)
	if err != nil {
		log.Printf("系统消息 %s 解析接收人群失败: %v", msg.ID, err)
		return 0, err
	}
	sent := DeliverSysMessage(msg, userIDs)
	return sent, repository.MarkSysMessageSent(msg.ID, sent)
}

// HandleDueSysMessages 定时扫描并发送到期的待发送消息（由 cron 每分钟调用）
func HandleDueSysMessages() {
	messages, err := repository.ListPendingSysMessages()
	if err != nil {
		log.Printf("扫描待发送系统消息失败: %v", err)
		return
	}
	for i := range messages {
		sent, err := SendSysMessage(&messages[i])
		if err != nil {
			log.Printf("定时发送系统消息 %s 失败: %v", messages[i].ID, err)
			continue
		}
		log.Printf("定时系统消息「%s」发送完成，共 %d 人", messages[i].Title, sent)
	}
}

// SendWelcomeMessage 新用户注册后发送欢迎消息
func SendWelcomeMessage(userID string) {
	notification := model.Notification{
		UserID:    userID,
		Type:      4, // 系统通知
		Content:   "欢迎您加入邻刻走 · LinkGo！在这里发现精彩攻略、寻找旅行搭子、记录每一次出行，开启你的旅程吧！",
		Title:     "欢迎加入邻刻走 · LinkGo",
		IsRead:    0,
		CreatedAt: time.Now(),
	}
	if err := database.DB.Create(&notification).Error; err != nil {
		log.Printf("发送新用户欢迎消息失败(user=%s): %v", userID, err)
		return
	}
	// 实时推送：新用户若在线（如刚注册登录），立即收到欢迎消息提醒
	ws.WsHub.PushToUser(userID, map[string]interface{}{
		"action":  "new_notification",
		"type":    4, // 系统通知
		"title":   notification.Title,
		"content": notification.Content,
	})
}

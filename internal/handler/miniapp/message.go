package miniapp

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/database"
	"travel-server/pkg/response"
)

// MessageVO 消息视图对象（含发送者信息）
type MessageVO struct {
	ID         string    `json:"id"`
	FromUserID string    `json:"fromUserId"`
	ToUserID   string    `json:"toUserId"`
	Content    string    `json:"content"`
	Type       int       `json:"type"`
	IsRead     int       `json:"isRead"`
	CreatedAt  time.Time `json:"createdAt"`
	AvatarURL  string    `json:"avatarUrl"`
	Nickname   string    `json:"nickname"`
}

// GetMessageList 获取与指定用户的私聊记录（分页，时间正序）
// @Summary 获取消息列表
// @Security BearerAuth
// @Tags 小程序-消息
// @Param targetUserId query string true "对方用户ID"
// @Param page query int false "页码（从最新一页开始）"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]MessageVO,total=int64}}
// @Router /api/v1/message/list [get]
func GetMessageList(c *gin.Context) {
	uid := c.MustGet("userID").(string)
	targetID := c.Query("targetUserId")
	if targetID == "" {
		response.Fail(c, 400, "缺少targetUserId")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	msgs, total, err := repository.GetMessagesBetweenUsers(uid, targetID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取消息失败")
		return
	}

	// 标记对方发给当前用户的未读消息为已读
	_ = repository.MarkMessagesAsRead(targetID, uid)

	// 收集所有发送者ID
	userIDs := make([]string, 0)
	seen := make(map[string]bool)
	for _, msg := range msgs {
		if !seen[msg.FromUserID] {
			seen[msg.FromUserID] = true
			userIDs = append(userIDs, msg.FromUserID)
		}
	}

	// 批量查询用户信息
	var users []model.User
	database.DB.Select("id, nickname, avatar_url").Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[string]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// 构造带用户信息的响应
	result := make([]MessageVO, 0, len(msgs))
	for _, msg := range msgs {
		u := userMap[msg.FromUserID]
		result = append(result, MessageVO{
			ID:         msg.ID,
			FromUserID: msg.FromUserID,
			ToUserID:   msg.ToUserID,
			Content:    msg.Content,
			Type:       msg.Type,
			IsRead:     msg.IsRead,
			CreatedAt:  msg.CreatedAt,
			AvatarURL:  u.AvatarURL,
			Nickname:   u.Nickname,
		})
	}

	response.Success(c, gin.H{"list": result, "total": total})
}

// SendMessage 发送私聊消息
// @Summary 发送消息
// @Security BearerAuth
// @Tags 小程序-消息
// @Param body body object{toUserId=string,content=string} true "消息内容"
// @Success 200 {object} response.Response
// @Router /api/v1/message/send [post]
func SendMessage(c *gin.Context) {
	var req struct {
		ToUserID string `json:"toUserId" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	msg := model.Message{
		FromUserID: c.MustGet("userID").(string),
		ToUserID:   req.ToUserID,
		Content:    req.Content,
		Type:       1, // 私聊
	}
	if err := repository.CreateMessage(&msg); err != nil {
		response.Fail(c, 500, "发送失败")
		return
	}
	response.Success(c, msg)
}

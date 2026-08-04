package miniapp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// GetMyConversations 获取统一会话列表（系统消息+私聊+群聊，按最后消息时间倒序）
// @Summary 会话列表
// @Security BearerAuth
// @Tags 小程序-群聊
// @Success 200 {object} response.Response{data=[]repository.ChatItemVO}
// @Router /api/v1/conversation/list [get]
func GetMyConversations(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	list, err := repository.GetChatList(userID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	if list == nil {
		list = []repository.ChatItemVO{}
	}
	response.Success(c, list)
}

// GetConversationDetail 群聊详情（群信息+成员列表）
// @Summary 群聊详情
// @Security BearerAuth
// @Tags 小程序-群聊
// @Param id path string true "群聊ID"
// @Success 200 {object} response.Response{data=object{id=string,partnerId=string,name=string,ownerId=string,isOwner=bool,isMember=bool,members=array}}
// @Router /api/v1/conversation/{id} [get]
func GetConversationDetail(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	conv, err := repository.GetConversationByID(c.Param("id"))
	if err != nil {
		response.Fail(c, 404, "群聊不存在")
		return
	}
	members, err := repository.GetConversationMembers(conv.ID)
	if err != nil {
		response.Fail(c, 500, "获取成员失败")
		return
	}
	response.Success(c, gin.H{
		"id":        conv.ID,
		"partnerId": conv.PartnerID,
		"name":      conv.Name,
		"ownerId":   conv.OwnerID,
		"isOwner":   conv.OwnerID == userID,
		"isMember":  repository.IsConversationMember(conv.ID, userID),
		"members":   members,
	})
}

// GetGroupMessages 群聊消息列表（分页，时间正序）
// @Summary 群聊消息列表
// @Security BearerAuth
// @Tags 小程序-群聊
// @Param id path string true "群聊ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]repository.ConversationMessageVO,total=int64}}
// @Router /api/v1/conversation/{id}/messages [get]
func GetGroupMessages(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	convID := c.Param("id")
	// 非成员不可查看群消息
	if !repository.IsConversationMember(convID, userID) {
		response.Fail(c, 403, "非群成员")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	list, total, err := repository.GetConversationMessages(convID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取消息失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// SendGroupMessage 发送群聊消息
// @Summary 发送群聊消息
// @Security BearerAuth
// @Tags 小程序-群聊
// @Param id path string true "群聊ID"
// @Param body body object{content=string} true "消息内容"
// @Success 200 {object} response.Response
// @Router /api/v1/conversation/{id}/message [post]
func SendGroupMessage(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	convID := c.Param("id")
	// 非成员不可发言
	if !repository.IsConversationMember(convID, userID) {
		response.Fail(c, 403, "非群成员")
		return
	}
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	msg := model.ConversationMessage{
		ConversationID: convID,
		FromUserID:     userID,
		Content:        req.Content,
	}
	if err := repository.CreateConversationMessage(&msg); err != nil {
		response.Fail(c, 500, "发送失败")
		return
	}
	response.Success(c, msg)
}

// KickConversationMember 踢出群成员（仅群主可操作）
// @Summary 踢出群成员
// @Security BearerAuth
// @Tags 小程序-群聊
// @Param id path string true "群聊ID"
// @Param body body object{userId=string} true "被踢成员用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/conversation/{id}/kick [put]
func KickConversationMember(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	convID := c.Param("id")
	var req struct {
		UserID string `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	conv, err := repository.GetConversationByID(convID)
	if err != nil {
		response.Fail(c, 404, "群聊不存在")
		return
	}
	// 仅群主可踢人
	if conv.OwnerID != userID {
		response.Fail(c, 403, "仅群主可踢人")
		return
	}
	// 不能踢群主自己
	if req.UserID == conv.OwnerID {
		response.Fail(c, 400, "不能踢出群主")
		return
	}
	if err := repository.KickConversationMember(convID, req.UserID); err != nil {
		response.Fail(c, 500, "踢出失败")
		return
	}
	response.Success(c, nil)
}

package admin

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/internal/service"
	"travel-server/pkg/database"
	"travel-server/pkg/response"
)

// 接收人群名称映射
var sysMessageTargetNames = map[string]string{
	"all":   "全部用户",
	"users": "指定用户",
	"group": "用户分组",
}

// 用户分组名称映射
var sysMessageGroupNames = map[string]string{
	"normal": "普通用户",
	"leader": "领队",
	"admin":  "管理员",
}

// CreateSysMessage 创建系统消息（立即发送 / 定时发送）
// @Summary 发送系统消息
// @Description 创建系统消息记录并推送：sendTime 为空或早于当前时间则立即发送，否则定时发送
// @Security BearerAuth
// @Tags 后台-消息管理
// @Accept json
// @Produce json
// @Param body body object{title=string,content=string,linkUrl=string,targetType=string,targetUserIds=[]string,targetGroup=string,sendTime=string} true "系统消息内容"
// @Success 200 {object} response.Response{data=model.SysMessage}
// @Router /api/v1/admin/sys-message [post]
func CreateSysMessage(c *gin.Context) {
	var req struct {
		Title         string   `json:"title" binding:"required"`
		Content       string   `json:"content" binding:"required"`
		LinkURL       string   `json:"linkUrl"`
		TargetType    string   `json:"targetType" binding:"required"`
		TargetUserIDs []string `json:"targetUserIds"`
		TargetGroup   string   `json:"targetGroup"`
		SendTime      string   `json:"sendTime"` // RFC3339，为空表示立即发送
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.TargetType != "all" && req.TargetType != "users" && req.TargetType != "group" {
		response.Fail(c, 400, "接收人群无效")
		return
	}
	if req.TargetType == "users" && len(req.TargetUserIDs) == 0 {
		response.Fail(c, 400, "请选择指定用户")
		return
	}
	if req.TargetType == "group" {
		if _, ok := sysMessageGroupNames[req.TargetGroup]; !ok {
			response.Fail(c, 400, "用户分组无效")
			return
		}
	}

	// 解析发送时间：缺省或早于当前 → 立即发送
	sendTime := time.Now()
	if req.SendTime != "" {
		t, err := time.Parse(time.RFC3339, req.SendTime)
		if err != nil {
			response.Fail(c, 400, "发送时间格式无效")
			return
		}
		sendTime = t
	}

	userIDsJSON := ""
	if req.TargetType == "users" {
		bytes, _ := json.Marshal(req.TargetUserIDs)
		userIDsJSON = string(bytes)
	}

	adminID, _ := c.Get("adminUserID")
	msg := &model.SysMessage{
		Title:         req.Title,
		Content:       req.Content,
		LinkURL:       req.LinkURL,
		TargetType:    req.TargetType,
		TargetUserIDs: userIDsJSON,
		TargetGroup:   req.TargetGroup,
		Status:        0,
		SendTime:      sendTime,
		OperatorID:    adminID.(string),
	}
	if err := repository.CreateSysMessage(msg); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}

	// 立即发送：sendTime 不晚于当前时间时直接推送
	if !sendTime.After(time.Now()) {
		if _, err := service.SendSysMessage(msg); err != nil {
			response.Fail(c, 500, "发送失败")
			return
		}
	}
	response.Success(c, msg)
}

// ListSysMessages 系统消息列表（分页，支持状态筛选）
// @Summary 系统消息列表
// @Security BearerAuth
// @Tags 后台-消息管理
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param status query int false "发送状态(0待发送 1已发送 2已取消)"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/sys-messages [get]
func ListSysMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	var status *int
	if s, err := strconv.Atoi(c.DefaultQuery("status", "-1")); err == nil && s >= 0 {
		status = &s
	}
	list, total, err := repository.ListSysMessages(page, pageSize, status)
	if err != nil {
		response.Fail(c, 500, "获取列表失败")
		return
	}

	// 批量查操作管理员昵称
	adminIDs := make([]string, 0, len(list))
	for _, m := range list {
		if m.OperatorID != "" {
			adminIDs = append(adminIDs, m.OperatorID)
		}
	}
	adminMap := make(map[string]string)
	if len(adminIDs) > 0 {
		var admins []model.AdminUser
		database.DB.Select("id, username").Where("id IN ?", adminIDs).Find(&admins)
		for _, a := range admins {
			adminMap[a.ID] = a.Username
		}
	}

	type item struct {
		model.SysMessage
		OperatorName string `json:"operatorName"` // 操作管理员用户名
		TargetName   string `json:"targetName"`   // 接收人群名称
		GroupName    string `json:"groupName"`    // 用户分组名称
	}
	result := make([]item, 0, len(list))
	for _, m := range list {
		it := item{SysMessage: m}
		it.OperatorName = adminMap[m.OperatorID]
		it.TargetName = sysMessageTargetNames[m.TargetType]
		it.GroupName = sysMessageGroupNames[m.TargetGroup]
		result = append(result, it)
	}
	response.Success(c, gin.H{"list": result, "total": total})
}

// CancelSysMessage 取消定时发送（仅待发送状态可取消）
// @Summary 取消定时发送
// @Security BearerAuth
// @Tags 后台-消息管理
// @Param id path string true "消息ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/sys-message/{id}/cancel [put]
func CancelSysMessage(c *gin.Context) {
	id := c.Param("id")
	if _, err := repository.GetSysMessageByID(id); err != nil {
		response.Fail(c, 404, "消息不存在")
		return
	}
	ok, err := repository.CancelSysMessage(id)
	if err != nil {
		response.Fail(c, 500, "取消失败")
		return
	}
	if !ok {
		response.Fail(c, 400, "仅待发送的消息可取消")
		return
	}
	response.Success(c, nil)
}

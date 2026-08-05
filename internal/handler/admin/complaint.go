package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/database"
	"travel-server/pkg/response"
)

// 投诉对象类型名称映射
var complaintTargetNames = map[string]string{
	"user": "用户", "guide": "攻略", "trip": "行程", "partner": "搭子", "comment": "评论", "other": "其他",
}

// ListComplaints 投诉列表（分页，支持状态/对象类型筛选）
// @Summary 投诉列表
// @Security BearerAuth
// @Tags 后台-投诉
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param status query int false "状态(0待处理 1已处理 2已驳回)"
// @Param targetType query string false "对象类型(user/guide/trip/partner/comment/other)"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/complaints [get]
func ListComplaints(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	targetType := c.Query("targetType")
	var status *int
	if s, err := strconv.Atoi(c.DefaultQuery("status", "-1")); err == nil && s >= 0 {
		status = &s
	}
	list, total, err := repository.ListComplaints(page, pageSize, status, targetType)
	if err != nil {
		response.Fail(c, 500, "获取投诉列表失败")
		return
	}

	// 批量收集投诉人ID与被投诉对象ID
	userIDs := make(map[string]struct{})
	objIDs := make(map[string]map[string]struct{}) // targetType -> ids
	for _, comp := range list {
		if comp.UserID != "" {
			userIDs[comp.UserID] = struct{}{}
		}
		if comp.TargetID == "" {
			continue
		}
		if objIDs[comp.TargetType] == nil {
			objIDs[comp.TargetType] = make(map[string]struct{})
		}
		objIDs[comp.TargetType][comp.TargetID] = struct{}{}
	}

	// 批量查投诉人
	userMap := make(map[string]model.User)
	if len(userIDs) > 0 {
		ids := make([]string, 0, len(userIDs))
		for id := range userIDs {
			ids = append(ids, id)
		}
		var users []model.User
		database.DB.Select("id, nickname, avatar_url").Where("id IN ?", ids).Find(&users)
		for i := range users {
			userMap[users[i].ID] = users[i]
		}
	}

	// 批量查被投诉对象摘要（标题/内容）
	targetMap := make(map[string]map[string]string) // targetType -> id -> 摘要
	for t, ids := range objIDs {
		idList := make([]string, 0, len(ids))
		for id := range ids {
			idList = append(idList, id)
		}
		targetMap[t] = make(map[string]string)
		switch t {
		case "user":
			var users []model.User
			database.DB.Select("id, nickname").Where("id IN ?", idList).Find(&users)
			for _, u := range users {
				targetMap[t][u.ID] = u.Nickname
			}
		case "guide":
			var objs []struct{ ID, Title string }
			database.DB.Model(&model.Guide{}).Select("id, title").Where("id IN ?", idList).Find(&objs)
			for _, o := range objs {
				targetMap[t][o.ID] = o.Title
			}
		case "trip":
			var objs []struct{ ID, Title string }
			database.DB.Model(&model.Trip{}).Select("id, title").Where("id IN ?", idList).Find(&objs)
			for _, o := range objs {
				targetMap[t][o.ID] = o.Title
			}
		case "partner":
			var objs []struct{ ID, Title string }
			database.DB.Model(&model.Partner{}).Select("id, title").Where("id IN ?", idList).Find(&objs)
			for _, o := range objs {
				targetMap[t][o.ID] = o.Title
			}
		case "comment":
			var objs []struct{ ID, Content string }
			database.DB.Model(&model.Comment{}).Select("id, content").Where("id IN ?", idList).Find(&objs)
			for _, o := range objs {
				targetMap[t][o.ID] = o.Content
			}
		}
	}

	type item struct {
		model.Complaint
		UserName   string `json:"userName"`   // 投诉人昵称
		AvatarURL  string `json:"avatarUrl"`  // 投诉人头像
		TargetName string `json:"targetName"` // 被投诉对象摘要
	}
	result := make([]item, 0, len(list))
	for _, comp := range list {
		it := item{Complaint: comp}
		if u, ok := userMap[comp.UserID]; ok {
			it.UserName = u.Nickname
			it.AvatarURL = u.AvatarURL
		}
		if m, ok := targetMap[comp.TargetType]; ok {
			it.TargetName = m[comp.TargetID]
		}
		result = append(result, it)
	}
	response.Success(c, gin.H{"list": result, "total": total})
}

// HandleComplaint 处理投诉（已处理/驳回，附处理备注与回复）
// @Summary 处理投诉
// @Security BearerAuth
// @Tags 后台-投诉
// @Param id path string true "投诉ID"
// @Param body body object{status=int,handleNote=string,reply=string} true "状态(1已处理 2已驳回)、备注与回复"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/complaint/{id}/status [put]
func HandleComplaint(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status     int    `json:"status"`
		HandleNote string `json:"handleNote"`
		Reply      string `json:"reply"` // 回复内容（小程序可见）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Status != 1 && req.Status != 2 {
		response.Fail(c, 400, "无效状态值")
		return
	}
	if _, err := repository.GetComplaintByID(id); err != nil {
		response.Fail(c, 404, "投诉不存在")
		return
	}
	if err := repository.UpdateComplaintStatus(id, req.Status, req.HandleNote, req.Reply); err != nil {
		response.Fail(c, 500, "处理失败")
		return
	}
	response.Success(c, nil)
}

// DeleteComplaint 删除投诉
// @Summary 删除投诉
// @Security BearerAuth
// @Tags 后台-投诉
// @Param id path string true "投诉ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/complaint/{id} [delete]
func DeleteComplaint(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DeleteComplaint(id); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

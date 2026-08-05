package miniapp

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
	"travel-server/pkg/snowflake"
)

// 合法投诉对象类型
var validComplaintTargets = map[string]bool{
	"user": true, "guide": true, "trip": true, "partner": true, "comment": true, "other": true,
}

// SubmitComplaint 提交投诉
// @Summary 提交投诉
// @Security BearerAuth
// @Tags 小程序-投诉
// @Param body body object{targetType=string,targetId=string,reason=string,content=string} true "投诉信息"
// @Success 200 {object} response.Response
// @Router /api/v1/complaint [post]
func SubmitComplaint(c *gin.Context) {
	var req struct {
		TargetType string   `json:"targetType"` // user/guide/trip/partner/comment/other
		TargetID   string   `json:"targetId"`   // 被投诉对象ID（other 时可不填）
		Reason     string   `json:"reason"`     // 投诉原因
		Content    string   `json:"content"`    // 详细描述
		Images     []string `json:"images"`     // 图片URL（最多9张）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	req.TargetType = strings.TrimSpace(req.TargetType)
	if !validComplaintTargets[req.TargetType] {
		response.Fail(c, 400, "无效的投诉对象类型")
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		response.Fail(c, 400, "请选择投诉原因")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if len([]rune(req.Content)) < 5 {
		response.Fail(c, 400, "请至少填写5个字的问题描述")
		return
	}
	if len([]rune(req.Content)) > 500 {
		response.Fail(c, 400, "问题描述不能超过500字")
		return
	}
	// 图片校验（最多9张）
	images := make([]string, 0, len(req.Images))
	for _, img := range req.Images {
		img = strings.TrimSpace(img)
		if img != "" {
			images = append(images, img)
		}
	}
	if len(images) > 9 {
		response.Fail(c, 400, "最多上传9张图片")
		return
	}
	imagesJSON, _ := json.Marshal(images)
	userID := c.MustGet("userID").(string)
	complaint := model.Complaint{
		ID:         snowflake.GenerateID(),
		UserID:     userID,
		TargetType: req.TargetType,
		TargetID:   strings.TrimSpace(req.TargetID),
		Reason:     req.Reason,
		Content:    req.Content,
		Images:     string(imagesJSON),
		Status:     0,
	}
	if err := repository.CreateComplaint(&complaint); err != nil {
		response.Fail(c, 500, "提交失败，请稍后重试")
		return
	}
	response.Success(c, gin.H{"id": complaint.ID})
}

// ListComplaints 我的投诉列表（分页）
// @Summary 我的投诉列表
// @Security BearerAuth
// @Tags 小程序-投诉
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]complaintItem,total=int}}
// @Router /api/v1/complaint/list [get]
func ListComplaints(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := repository.ListUserComplaints(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取列表失败")
		return
	}
	result := make([]complaintItem, 0, len(list))
	for i := range list {
		result = append(result, toComplaintItem(&list[i]))
	}
	response.Success(c, gin.H{"list": result, "total": total})
}

// GetComplaintDetail 我的投诉详情（仅本人）
// @Summary 我的投诉详情
// @Security BearerAuth
// @Tags 小程序-投诉
// @Param id path string true "投诉ID"
// @Success 200 {object} response.Response{data=complaintItem}
// @Router /api/v1/complaint/{id} [get]
func GetComplaintDetail(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	complaint, err := repository.GetUserComplaint(c.Param("id"), userID)
	if err != nil {
		response.Fail(c, 404, "投诉不存在")
		return
	}
	response.Success(c, toComplaintItem(complaint))
}

// complaintItem 投诉条目（images 解析为数组返回）
type complaintItem struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	TargetType string     `json:"targetType"`
	TargetID   string     `json:"targetId"`
	Reason     string     `json:"reason"`
	Content    string     `json:"content"`
	Images     []string   `json:"images"`
	Status     int        `json:"status"`
	HandleNote string     `json:"handleNote"`
	Reply      string     `json:"reply"`
	HandledAt  *time.Time `json:"handledAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func toComplaintItem(c *model.Complaint) complaintItem {
	var images []string
	if c.Images != "" {
		_ = json.Unmarshal([]byte(c.Images), &images)
	}
	if images == nil {
		images = []string{}
	}
	return complaintItem{
		ID:         c.ID,
		UserID:     c.UserID,
		TargetType: c.TargetType,
		TargetID:   c.TargetID,
		Reason:     c.Reason,
		Content:    c.Content,
		Images:     images,
		Status:     c.Status,
		HandleNote: c.HandleNote,
		Reply:      c.Reply,
		HandledAt:  c.HandledAt,
		CreatedAt:  c.CreatedAt,
	}
}

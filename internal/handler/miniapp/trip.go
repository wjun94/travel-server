package miniapp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"travel-server/internal/ai"
	"travel-server/internal/middleware"
	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// AIGenerateTrip 调用 DeepSeek 智能生成行程
// @Summary AI生成行程
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body object{destination=string,days=int,tags=[]string} true "生成参数"
// @Success 200 {object} response.Response{data=model.Trip}
// @Router /api/v1/trip/ai-generate [post]
func AIGenerateTrip(c *gin.Context) {
	var req struct {
		Destination string   `json:"destination"`
		Days        int      `json:"days"`
		Tags        []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	uid := c.GetUint("userID")

	prompt := fmt.Sprintf(ai.TripPrompt, req.Destination, req.Days, strings.Join(req.Tags, "、"))
	result, err := ai.Chat(prompt)
	if err != nil {
		response.Fail(c, 500, "AI生成失败")
		return
	}

	// 解析返回的 JSON
	var plan struct {
		Days []model.DailyPlan `json:"days"`
	}
	if err := json.Unmarshal([]byte(result), &plan); err != nil {
		response.Fail(c, 500, "AI返回格式异常")
		return
	}

	trip := model.Trip{
		UserID:      uid,
		Destination: req.Destination,
		Days:        req.Days,
		DailyPlans:  model.JSONString(result),
	}
	if err := repository.CreateTrip(&trip); err != nil {
		response.Fail(c, 500, "保存失败")
		return
	}
	response.Success(c, trip)
}

// CreateTrip 手动创建行程
// @Summary 创建手动行程
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body model.Trip true "行程信息"
// @Success 200 {object} response.Response
// @Router /api/v1/trip [post]
func CreateTrip(c *gin.Context) {
	var trip model.Trip
	if err := c.ShouldBindJSON(&trip); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	trip.UserID = c.GetUint("userID")
	if err := repository.CreateTrip(&trip); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}
	response.Success(c, trip)
}

// GetTrip 获取行程详情
// @Summary 获取行程详情
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path int true "行程ID"
// @Success 200 {object} response.Response{data=model.Trip}
// @Router /api/v1/trip/{id} [get]
func GetTrip(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	trip, err := repository.GetTripByID(uint(id))
	if err != nil {
		response.Fail(c, 404, "行程不存在")
		return
	}
	response.Success(c, trip)
}

// UpdateTrip 协同编辑行程
// @Summary 更新行程
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path int true "行程ID"
// @Param body body object{daily_plans=object} true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/{id} [put]
func UpdateTrip(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	trip, err := repository.GetTripByID(uint(id))
	if err != nil {
		response.Fail(c, 404, "行程不存在")
		return
	}
	// 权限检查：创建者或协作者（编辑权限）
	userID := c.GetUint("userID")
	if trip.UserID != userID {
		hasPerm := false
		for _, col := range trip.Collaborators {
			if col.UserID == userID && col.Permission == 1 {
				hasPerm = true
				break
			}
		}
		if !hasPerm {
			response.Fail(c, 403, "无编辑权限")
			return
		}
	}

	var req struct {
		DailyPlans json.RawMessage `json:"daily_plans"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}

	trip.DailyPlans = model.JSONString(req.DailyPlans)
	if err := repository.UpdateTrip(trip); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, trip)
}

// InviteCollaborator 生成邀请链接
// @Summary 邀请好友协同编辑
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path int true "行程ID"
// @Success 200 {object} response.Response{data=object{invite_url=string}}
// @Router /api/v1/trip/{id}/invite [post]
func InviteCollaborator(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// 简单生成邀请 token（实际应加密含 tripID 的临时凭据）
	inviteToken, _ := middleware.GenerateMiniAppToken(c.GetUint("userID"))
	inviteUrl := fmt.Sprintf("/pages/trip/detail?id=%d&token=%s", id, inviteToken)
	response.Success(c, gin.H{"inviteUrl": inviteUrl})
}

package miniapp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/ai"
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
	uid := c.MustGet("userID").(string)

	prompt := fmt.Sprintf(ai.TripPrompt, req.Destination, req.Days, "")
	result, err := ai.Chat(prompt)
	if err != nil {
		response.Fail(c, 500, "AI生成失败")
		return
	}

	// 解析 AI 返回的行程日数据
	var aiResult struct {
		Days []struct {
			Day   int `json:"day"`
			Items []struct {
				Time     string `json:"time"`
				Name     string `json:"name"`
				Type     string `json:"type"`
				Duration string `json:"duration"`
			} `json:"items"`
		} `json:"days"`
	}
	if err := json.Unmarshal([]byte(result), &aiResult); err != nil {
		response.Fail(c, 500, "AI返回格式异常")
		return
	}

	// 创建行程
	trip := model.Trip{
		UserID:       uid,
		Destinations: []string{req.Destination},
		Status:       1,
	}
	if err := repository.CreateTrip(&trip); err != nil {
		response.Fail(c, 500, "保存失败")
		return
	}

	// 创建行程日及行程项
	for _, d := range aiResult.Days {
		day := model.TripDay{
			TripID:    trip.ID,
			DayNumber: d.Day,
		}
		repository.CreateTripDay(&day)
		for _, item := range d.Items {
			tripItem := model.TripItem{
				TripDayID:   day.ID,
				StartTime:   item.Time,
				SectionType: item.Type,
				Title:       item.Name,
			}
			repository.CreateTripItem(&tripItem)
		}
	}

	// 重新加载完整数据
	fullTrip, _ := repository.GetTripByID(trip.ID)
	response.Success(c, fullTrip)
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
	trip.UserID = c.MustGet("userID").(string)
	if err := repository.CreateTrip(&trip); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}
	response.Success(c, trip)
}

// GetTrip 获取行程详情
// @Summary 获取行程详情
// @Tags 小程序-行程
// @Param id path string true "行程ID"
// @Success 200 {object} response.Response{data=model.Trip}
// @Router /api/v1/trip/{id} [get]
func GetTrip(c *gin.Context) {
	id := c.Param("id")
	trip, err := repository.GetTripByID(id)
	if err != nil {
		response.Fail(c, 404, "行程不存在")
		return
	}
	response.Success(c, trip)
}

// GetMyTrips 我的行程列表
// @Summary 我的行程列表
// @Security BearerAuth
// @Tags 小程序-行程
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.Trip,total=int}}
// @Router /api/v1/my/trips [get]
func GetMyTrips(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	trips, total, err := repository.ListUserTrips(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": trips, "total": total})
}

// UpdateTrip 更新行程基本信息
// @Summary 更新行程
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程ID"
// @Param body body object true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/{id} [put]
func UpdateTrip(c *gin.Context) {
	id := c.Param("id")
	trip, err := repository.GetTripByID(id)
	if err != nil {
		response.Fail(c, 404, "行程不存在")
		return
	}
	// 权限检查：仅创建者可编辑
	userID := c.MustGet("userID").(string)
	if trip.UserID != userID {
		response.Fail(c, 403, "无编辑权限")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	// 过滤不可更新的字段
	delete(updates, "id")
	delete(updates, "user_id")
	delete(updates, "created_at")
	delete(updates, "view_count")
	delete(updates, "like_count")
	delete(updates, "favorite_count")

	if err := repository.UpdateTrip(id, updates); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// ==================== TripDay 行程日 ====================

// AddTripDay 添加行程日
// @Summary 添加行程日
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body model.TripDay true "行程日信息"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/day [post]
func AddTripDay(c *gin.Context) {
	var day model.TripDay
	if err := c.ShouldBindJSON(&day); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.CreateTripDay(&day); err != nil {
		response.Fail(c, 500, "添加失败")
		return
	}
	response.Success(c, day)
}

// UpdateTripDay 更新行程日
// @Summary 更新行程日
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程日ID"
// @Param body body object true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/day/{id} [put]
func UpdateTripDay(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.UpdateTripDay(id, updates); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// DeleteTripDay 删除行程日
// @Summary 删除行程日
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程日ID"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/day/{id} [delete]
func DeleteTripDay(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DeleteTripDay(id); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

// ==================== TripItem 行程项 ====================

// AddTripItem 添加行程项
// @Summary 添加行程项
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body model.TripItem true "行程项信息"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/item [post]
func AddTripItem(c *gin.Context) {
	var item model.TripItem
	if err := c.ShouldBindJSON(&item); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if item.SectionType != "" && !model.ValidSectionTypes[item.SectionType] {
		response.Fail(c, 400, "无效的行程项类型，可选：transport/hotel/attraction/food/shopping/tips")
		return
	}
	if err := repository.CreateTripItem(&item); err != nil {
		response.Fail(c, 500, "添加失败")
		return
	}
	response.Success(c, item)
}

// UpdateTripItem 更新行程项
// @Summary 更新行程项
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程项ID"
// @Param body body object true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/item/{id} [put]
func UpdateTripItem(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if sectionType, ok := updates["sectionType"].(string); ok && sectionType != "" && !model.ValidSectionTypes[sectionType] {
		response.Fail(c, 400, "无效的行程项类型，可选：transport/hotel/attraction/food/shopping/tips")
		return
	}
	if err := repository.UpdateTripItem(id, updates); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// DeleteTripItem 删除行程项
// @Summary 删除行程项
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程项ID"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/item/{id} [delete]
func DeleteTripItem(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DeleteTripItem(id); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

// ==================== TripMember 同行者 ====================

// InviteMember 邀请同行者
// @Summary 邀请同行者
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body model.TripMember true "同行者信息"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/member [post]
func InviteMember(c *gin.Context) {
	var member model.TripMember
	if err := c.ShouldBindJSON(&member); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.AddTripMember(&member); err != nil {
		response.Fail(c, 500, "邀请失败")
		return
	}
	response.Success(c, member)
}

// RemoveMember 移除同行者
// @Summary 移除同行者
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "同行者记录ID"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/member/{id} [delete]
func RemoveMember(c *gin.Context) {
	id := c.Param("id")
	if err := repository.RemoveTripMember(id); err != nil {
		response.Fail(c, 500, "移除失败")
		return
	}
	response.Success(c, nil)
}

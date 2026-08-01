package miniapp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/ai"
	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/database"
	"travel-server/pkg/response"
)

// AIGenerateTrip 调用 DeepSeek 智能生成行程
// @Summary AI生成行程
// @Description 根据目的地和天数调用AI生成完整行程，自动填充标题/国家/省市/预算/概述等字段并保存（每日基础1次，邀请好友成功1人可额外+1次，超出返回400）
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body object{destination=string,days=int} true "生成参数：destination=目的地, days=旅行天数"
// @Success 200 {object} response.Response{data=model.Trip} "返回完整行程（含国家/省市/预算/概述及每日行程项）"
// @Router /api/v1/trip/ai-generate [post]
func AIGenerateTrip(c *gin.Context) {
	var req struct {
		Destination string `json:"destination" binding:"required"`
		Days        int    `json:"days" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	uid := c.MustGet("userID").(string)

	// 额度校验：今日基础1次 + 邀请成功奖励，超出拒绝
	inviteCount, _ := repository.CountTodayInviteSuccess(uid)
	tripUsed, _ := repository.CountTodayAITrips(uid)
	if int(tripUsed) >= 1+int(inviteCount) {
		response.Fail(c, 400, "今日AI生成次数已用完，邀请好友可额外获得次数")
		return
	}

	prompt := fmt.Sprintf(ai.TripPrompt, req.Destination, req.Days)
	result, err := ai.Chat(prompt)
	if err != nil {
		response.Fail(c, 500, "AI生成失败")
		return
	}

	// 解析 AI 返回的行程数据（含行程基础信息与每日行程）
	var aiResult struct {
		Title       string   `json:"title"`
		Countries   []string `json:"countries"`
		Provinces   []string `json:"provinces"`
		Cities      []string `json:"cities"`
		IsOverseas  int      `json:"isOverseas"`
		TotalBudget float64  `json:"totalBudget"`
		Summary     string   `json:"summary"`
		Days        []struct {
			Day   int `json:"day"`
			Items []struct {
				Time        string `json:"time"`
				Name        string `json:"name"`
				Type        string `json:"type"`
				Duration    string `json:"duration"`
				Address     string `json:"address"`
				Description string `json:"description"`
			} `json:"items"`
		} `json:"days"`
	}
	if err := json.Unmarshal([]byte(result), &aiResult); err != nil {
		response.Fail(c, 500, "AI返回格式异常")
		return
	}

	// 创建行程（填充AI返回的完整字段）
	trip := model.Trip{
		UserID:       uid,
		Title:        aiResult.Title,
		Countries:    aiResult.Countries,
		Provinces:    aiResult.Provinces,
		Cities:       aiResult.Cities,
		Destinations: []string{req.Destination},
		IsOverseas:   aiResult.IsOverseas,
		TotalBudget:  aiResult.TotalBudget,
		Summary:      aiResult.Summary,
		Status:       1, // 草稿
		IsAI:         1, // AI生成
	}
	// 兜底：AI未返回标题时用目的地+天数拼接
	if trip.Title == "" {
		trip.Title = fmt.Sprintf("%s%d日游", req.Destination, req.Days)
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
			Title:     fmt.Sprintf("第%d天", d.Day),
		}
		repository.CreateTripDay(&day)
		for _, item := range d.Items {
			tripItem := model.TripItem{
				TripDayID:   day.ID,
				StartTime:   item.Time,
				SectionType: item.Type,
				Title:       item.Name,
				Address:     item.Address,
				Description: item.Description,
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
	// 作者信息
	author, _ := repository.GetUserByID(trip.UserID)
	authorName := ""
	authorAvatar := ""
	if author != nil {
		authorName = author.Nickname
		authorAvatar = author.AvatarURL
	}
	response.Success(c, gin.H{
		"id":            trip.ID,
		"userId":        trip.UserID,
		"guideId":       trip.GuideID,
		"title":         trip.Title,
		"coverImage":    trip.CoverImage,
		"countries":     trip.Countries,
		"provinces":     trip.Provinces,
		"cities":        trip.Cities,
		"destinations":  trip.Destinations,
		"totalBudget":   trip.TotalBudget,
		"isOverseas":    trip.IsOverseas,
		"summary":       trip.Summary,
		"viewCount":     trip.ViewCount,
		"likeCount":     trip.LikeCount,
		"favoriteCount": trip.FavoriteCount,
		"status":        trip.Status,
		"isPublic":      trip.IsPublic,
		"createdAt":     trip.CreatedAt,
		"updatedAt":     trip.UpdatedAt,
		"authorName":    authorName,
		"authorAvatar":  authorAvatar,
		"isFollowed":    false,
		"isSelf":        true,
	})
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
		response.Fail(c, 500, "行程不存在")
		return
	}
	userID := c.MustGet("userID").(string)

	// 作者信息
	author, _ := repository.GetUserByID(trip.UserID)
	authorName := ""
	authorAvatar := ""
	if author != nil {
		authorName = author.Nickname
		authorAvatar = author.AvatarURL
	}
	// 关注状态
	followStatus, _ := repository.GetFollowStatus(userID, trip.UserID)
	isFollowed := followStatus == 1 || followStatus == 2

	// 收藏数、评论数、点赞状态
	favoriteCount := repository.GetTripFavoriteCount(id)
	commentCount := repository.GetTripCommentCount(id)
	isLiked := repository.IsFavorited(userID, id, "trip")

	// 将 trip 展开到顶层，不嵌套在单独的 trip 字段里
	result := gin.H{
		"id":            trip.ID,
		"userId":        trip.UserID,
		"guideId":       trip.GuideID,
		"title":         trip.Title,
		"coverImage":    trip.CoverImage,
		"countries":     trip.Countries,
		"provinces":     trip.Provinces,
		"cities":        trip.Cities,
		"destinations":  trip.Destinations,
		"totalBudget":   trip.TotalBudget,
		"isOverseas":    trip.IsOverseas,
		"summary":       trip.Summary,
		"viewCount":     trip.ViewCount,
		"likeCount":     trip.LikeCount,
		"favoriteCount": favoriteCount,
		"commentCount":  commentCount,
		"status":        trip.Status,
		"isPublic":      trip.IsPublic,
		"createdAt":     trip.CreatedAt,
		"updatedAt":     trip.UpdatedAt,
		"days":          trip.Days,
		"members":       trip.Members,
		"isLiked":       isLiked,
		"isFavorited":   isLiked,
		"authorName":    authorName,
		"authorAvatar":  authorAvatar,
		"isFollowed":    isFollowed,
		"isSelf":        userID == trip.UserID,
	}

	response.Success(c, result)
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

	// 批量查询行程的天数和行程项统计
	tripIDs := make([]string, len(trips))
	for i, t := range trips {
		tripIDs[i] = t.ID
	}

	type tripDayCount struct {
		TripID string
		Days   int
	}
	var tripDayCounts []tripDayCount
	if len(tripIDs) > 0 {
		database.DB.Model(&model.TripDay{}).
			Select("trip_id, COUNT(DISTINCT day_number) as days").
			Where("trip_id IN ?", tripIDs).
			Group("trip_id").
			Find(&tripDayCounts)
	}
	dayMap := make(map[string]int)
	for _, d := range tripDayCounts {
		dayMap[d.TripID] = d.Days
	}

	type tripSecCount struct {
		TripID string
		Count  int64
	}
	var tripSecCounts []tripSecCount
	if len(tripIDs) > 0 {
		database.DB.Table("trip_items ti").
			Select("td.trip_id, COUNT(ti.id) as count").
			Joins("LEFT JOIN trip_days td ON td.id = ti.trip_day_id").
			Where("td.trip_id IN ?", tripIDs).
			Group("td.trip_id").
			Find(&tripSecCounts)
	}
	secMap := make(map[string]int64)
	for _, s := range tripSecCounts {
		secMap[s.TripID] = s.Count
	}

	// 组装返回数据
	list := make([]gin.H, 0, len(trips))
	for _, t := range trips {
		list = append(list, gin.H{
			"id":            t.ID,
			"userId":        t.UserID,
			"guideId":       t.GuideID,
			"title":         t.Title,
			"coverImage":    t.CoverImage,
			"countries":     t.Countries,
			"provinces":     t.Provinces,
			"cities":        t.Cities,
			"destinations":  t.Destinations,
			"totalBudget":   t.TotalBudget,
			"isOverseas":    t.IsOverseas,
			"summary":       t.Summary,
			"viewCount":     t.ViewCount,
			"likeCount":     t.LikeCount,
			"favoriteCount": t.FavoriteCount,
			"tripDays":      dayMap[t.ID],
			"sectionCount":  secMap[t.ID],
			"status":        t.Status,
			"isPublic":      t.IsPublic,
			"createdAt":     t.CreatedAt,
			"updatedAt":     t.UpdatedAt,
		})
	}

	response.Success(c, gin.H{"list": list, "total": total})
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
